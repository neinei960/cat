#!/bin/bash
# MySQL Binlog 查询专家
# 用法: ./binlog-query.sh <操作描述>
# 例如:
#   ./binlog-query.sh "DELETE FROM customers"
#   ./binlog-query.sh "UPDATE pets SET avatar"
#   ./binlog-query.sh "月光"
#   ./binlog-query.sh customers 2026-03-25

SERVER="36.151.144.227"
SSHPASS="Gls6622821-"
DB_PASS="root123"

QUERY="$1"
DATE_FILTER="$2"

SSH="sshpass -p $SSHPASS ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@$SERVER"

if [ -z "$QUERY" ]; then
  cat << 'HELP'
╔══════════════════════════════════════════════╗
║       MySQL Binlog 查询专家                  ║
╚══════════════════════════════════════════════╝

用法: ./binlog-query.sh <搜索关键词> [日期]

示例:
  ./binlog-query.sh customers              # 搜索所有 customers 表操作
  ./binlog-query.sh 'DELETE FROM pets'      # 搜索删除宠物的操作
  ./binlog-query.sh '月光'                  # 搜索包含"月光"的操作
  ./binlog-query.sh customers 2026-03-25   # 只看某天的操作

其他命令:
  ./binlog-query.sh --logs                  # 列出所有 binlog 文件
  ./binlog-query.sh --recent [表名]         # 查看最近的操作
  ./binlog-query.sh --tables                # 列出被修改过的表
HELP
  exit 0
fi

# 列出 binlog 文件
if [ "$QUERY" = "--logs" ]; then
  echo "📋 Binlog 文件列表:"
  $SSH "mysql -u root -p$DB_PASS -e 'SHOW BINARY LOGS;' 2>/dev/null | grep -v Warning"
  echo ""
  echo "当前写入:"
  $SSH "mysql -u root -p$DB_PASS -e 'SHOW MASTER STATUS;' 2>/dev/null | grep -v Warning"
  exit 0
fi

# 查看最近被修改的表
if [ "$QUERY" = "--tables" ]; then
  echo "📊 最近 binlog 中被修改的表:"
  echo ""
  $SSH bash << 'REMOTE'
mysqlbinlog --no-defaults -v --base64-output=DECODE-ROWS /var/lib/mysql/binlog.000007 2>/dev/null | \
  grep -oP '(?<=### )(INSERT INTO|UPDATE|DELETE FROM) `\w+`\.`\w+`' | \
  sort | uniq -c | sort -rn | head -20
REMOTE
  exit 0
fi

# 查看某张表最近操作
if [ "$QUERY" = "--recent" ]; then
  TABLE="${2:-customers}"
  echo "🕐 表 $TABLE 最近操作:"
  echo ""
  $SSH bash -c "'mysqlbinlog --no-defaults -v --base64-output=DECODE-ROWS /var/lib/mysql/binlog.000007 2>/dev/null | grep -B2 -A10 \"\\\`$TABLE\\\`\" | tail -100'"
  exit 0
fi

# 主查询：在服务器上执行搜索
echo "🔍 搜索 binlog: \"$QUERY\""
[ -n "$DATE_FILTER" ] && echo "📅 日期过滤: $DATE_FILTER"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

$SSH bash -s "$QUERY" "$DATE_FILTER" << 'REMOTE_SCRIPT'
QUERY="$1"
DATE_FILTER="$2"
DB_PASS="root123"

BINLOGS=$(mysql -u root -p$DB_PASS -N -e 'SHOW BINARY LOGS;' 2>/dev/null | awk '{print $1}')

for BINLOG in $BINLOGS; do
  DATE_ARG=""
  if [ -n "$DATE_FILTER" ]; then
    DATE_ARG="--start-datetime=${DATE_FILTER}T00:00:00 --stop-datetime=${DATE_FILTER}T23:59:59"
  fi

  RESULTS=$(mysqlbinlog --no-defaults -v --base64-output=DECODE-ROWS $DATE_ARG /var/lib/mysql/$BINLOG 2>/dev/null | \
    grep -B5 -A15 "$QUERY" | head -300)

  if [ -n "$RESULTS" ]; then
    echo "📁 $BINLOG:"
    echo "$RESULTS"
    echo ""
    echo "---"
    echo ""
  fi
done

echo "✅ 搜索完成"
REMOTE_SCRIPT
