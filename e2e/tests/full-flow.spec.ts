import { test, expect } from '@playwright/test';
import mysql from 'mysql2/promise';

const BASE = 'http://localhost:8080/api/v1';

// --- DB helper ---
let db: mysql.Connection;

async function query(sql: string, params?: any[]) {
  if (!db) {
    db = await mysql.createConnection({
      host: '127.0.0.1', port: 3306,
      user: 'root', password: 'root123', database: 'petshop',
    });
  }
  const [rows] = await db.execute(sql, params);
  return rows as any[];
}

function authHeaders(token: string) {
  return { Authorization: `Bearer ${token}` };
}

function localDate(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

// --- shared state (file-level, survives across serial describes) ---
const state = {
  adminToken: '',
  staffToken: '',
  testCustomerId: 0,
  testPetId: 0,
  testOrderId: 0,
  testAppointmentId: 0,
};

async function ensureTokens(request: any) {
  if (!state.adminToken) {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    state.adminToken = (await res.json()).data.token;
  }
  if (!state.staffToken) {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138001', password: '123456' },
    });
    state.staffToken = (await res.json()).data.token;
  }
}

// ============================================================
// 1. 基础健康检查 & 登录
// ============================================================
test.describe.serial('1. 基础: 健康检查 & 登录', () => {
  test('1.1 健康检查', async ({ request }) => {
    const res = await request.get(`${BASE}/health`);
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.status).toBe('ok');
  });

  test('1.2 管理员登录 (admin)', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.token).toBeTruthy();
    expect(data.data.staff.role).toBe('admin');
    state.adminToken = data.data.token;
  });

  test('1.3 普通员工登录 (staff)', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138001', password: '123456' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.staff.role).toBe('staff');
    state.staffToken = data.data.token;
  });

  test('1.4 错误密码拒绝登录', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: 'wrong' },
    });
    const data = await res.json();
    expect(data.code).not.toBe(0);
  });

  test('1.5 无token访问被拒', async ({ request }) => {
    const res = await request.get(`${BASE}/b/shop`);
    expect(res.status()).toBe(401);
  });
});

// ============================================================
// 2. 客户管理
// ============================================================
test.describe.serial('2. 客户管理', () => {
  test('2.1 创建客户', async ({ request }) => {
    const phone = `139${Date.now().toString().slice(-8)}`;
    const res = await request.post(`${BASE}/b/customers`, {
      headers: authHeaders(state.adminToken),
      data: { nickname: 'E2E测试客户', phone, gender: 2, remark: 'Playwright自动创建' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.ID).toBeGreaterThan(0);
    state.testCustomerId = data.data.ID;

    // DB 验证
    const rows = await query('SELECT * FROM customers WHERE id = ?', [state.testCustomerId]);
    expect(rows.length).toBe(1);
    expect(rows[0].nickname).toBe('E2E测试客户');
  });

  test('2.2 查看客户列表', async ({ request }) => {
    const res = await request.get(`${BASE}/b/customers`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThan(0);
  });

  test('2.3 客户搜索 (by keyword)', async ({ request }) => {
    const res = await request.get(`${BASE}/b/customers?keyword=E2E测试`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((c: any) => c.ID === state.testCustomerId);
    expect(found).toBeTruthy();
  });

  test('2.4 编辑客户', async ({ request }) => {
    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}`, {
      headers: authHeaders(state.adminToken),
      data: { nickname: 'E2E测试客户-已修改', phone: '', gender: 2, remark: '已编辑', tags: '测试,VIP' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const rows = await query('SELECT nickname, remark FROM customers WHERE id = ?', [state.testCustomerId]);
    expect(rows[0].nickname).toBe('E2E测试客户-已修改');
  });
});

// ============================================================
// 3. 宠物管理 (含新功能: 手机号关联、搜索、性别)
// ============================================================
test.describe.serial('3. 宠物管理', () => {
  test('3.0 确保token', async ({ request }) => { await ensureTokens(request); });
  test('3.1 创建宠物 (通过owner_phone自动关联客户)', async ({ request }) => {
    // 用唯一手机号避免冲突
    const testPhone = `137${Date.now().toString().slice(-8)}`;
    await query('UPDATE customers SET phone = ? WHERE id = ?', [testPhone, state.testCustomerId]);

    const res = await request.post(`${BASE}/b/pets`, {
      headers: authHeaders(state.adminToken),
      data: {
        name: 'E2E测试猫',
        species: '猫',
        breed: '英短',
        gender: 2, // 妹妹
        weight: 4.5,
        fur_level: '短毛猫',
        owner_phone: testPhone,
        birth_date: '2024-06-15',
        personality: '神仙宝贝',
        aggression: '无',
        neutered: true,
        care_notes: '怕水，需轻柔操作',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.ID).toBeGreaterThan(0);
    state.testPetId = data.data.ID;

    // DB 验证: customer_id 应该被自动关联
    const rows = await query('SELECT customer_id, name, gender, neutered FROM pets WHERE id = ?', [state.testPetId]);
    expect(rows.length).toBe(1);
    // customer_id 应该被关联到我们的测试客户
    const custRow = await query('SELECT id FROM customers WHERE phone = ?', [testPhone]);
    expect(custRow.length).toBe(1);
    expect(rows[0].customer_id).toBe(custRow[0].id);
    expect(rows[0].gender).toBe(2);
    expect(rows[0].neutered).toBe(1);
  });

  test('3.2 创建宠物 (新手机号自动创建客户)', async ({ request }) => {
    const newPhone = '13700009999';
    const res = await request.post(`${BASE}/b/pets`, {
      headers: authHeaders(state.adminToken),
      data: {
        name: 'E2E自动关联猫',
        species: '猫',
        breed: '布偶',
        gender: 1,
        fur_level: '长毛猫',
        owner_phone: newPhone,
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    // DB 验证: 应该自动创建了客户
    const customers = await query('SELECT id FROM customers WHERE phone = ?', [newPhone]);
    expect(customers.length).toBe(1);

    const pet = await query('SELECT customer_id FROM pets WHERE id = ?', [data.data.ID]);
    expect(pet[0].customer_id).toBe(customers[0].id);

    // 清理
    await query('DELETE FROM pets WHERE id = ?', [data.data.ID]);
    await query('DELETE FROM customers WHERE id = ?', [customers[0].id]);
  });

  test('3.3 宠物列表 & 搜索 (按名称)', async ({ request }) => {
    const res = await request.get(`${BASE}/b/pets?keyword=E2E测试猫`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((p: any) => p.ID === state.testPetId);
    expect(found).toBeTruthy();
    expect(found.name).toBe('E2E测试猫');
  });

  test('3.4 宠物搜索 (按主人手机号)', async ({ request }) => {
    // 获取当前测试客户的手机号
    const custRows = await query('SELECT phone FROM customers WHERE id = ?', [state.testCustomerId]);
    const phone = custRows[0].phone;
    const res = await request.get(`${BASE}/b/pets?keyword=${phone}`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((p: any) => p.ID === state.testPetId);
    expect(found).toBeTruthy();
  });

  test('3.5 宠物搜索 (按主人昵称)', async ({ request }) => {
    const res = await request.get(`${BASE}/b/pets?keyword=E2E测试客户`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((p: any) => p.ID === state.testPetId);
    expect(found).toBeTruthy();
  });

  test('3.6 编辑宠物', async ({ request }) => {
    const res = await request.put(`${BASE}/b/pets/${state.testPetId}`, {
      headers: authHeaders(state.adminToken),
      data: {
        name: 'E2E测试猫-改名',
        species: '猫',
        breed: '英短蓝猫',
        gender: 2,
        weight: 5.0,
        fur_level: '短毛猫',
        owner_phone: (await query('SELECT phone FROM customers WHERE id = ?', [state.testCustomerId]))[0].phone,
        care_notes: '已更新注意事项',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const rows = await query('SELECT name, weight, care_notes FROM pets WHERE id = ?', [state.testPetId]);
    expect(rows[0].name).toBe('E2E测试猫-改名');
    expect(parseFloat(rows[0].weight)).toBe(5.0);
  });

  test('3.7 宠物列表按主人分组排序', async ({ request }) => {
    const res = await request.get(`${BASE}/b/pets?page_size=200`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    // 验证同主人的宠物相邻（customer_id 排序）
    const list = data.data.list;
    if (list.length > 2) {
      let lastCid = list[0].customer_id || 0;
      let seenCids = new Set<number>();
      let groupBroken = false;
      for (const pet of list) {
        const cid = pet.customer_id || 0;
        if (cid !== lastCid) {
          if (seenCids.has(cid)) {
            groupBroken = true;
            break;
          }
          seenCids.add(lastCid);
          lastCid = cid;
        }
      }
      expect(groupBroken).toBe(false);
    }
  });
});

// ============================================================
// 4. 毛发类别管理

// ============================================================
// 5. 订单管理 (含搜索功能)
// ============================================================
test.describe.serial('5. 订单管理', () => {
  test('5.0 确保token', async ({ request }) => { await ensureTokens(request); });
  test('5.1 快速开单', async ({ request }) => {
    // 获取服务列表找到第一个服务
    const svcRes = await request.get(`${BASE}/b/services`, {
      headers: authHeaders(state.adminToken),
    });
    const svcData = await svcRes.json();
    const serviceId = svcData.data.list[0]?.ID;
    expect(serviceId).toBeGreaterThan(0);

    const res = await request.post(`${BASE}/b/orders`, {
      headers: authHeaders(state.adminToken),
      data: {
        pet_id: state.testPetId,
        service_id: serviceId,
        remark: 'E2E测试订单',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.order_no).toBeTruthy();
    expect(data.data.pay_amount).toBeGreaterThan(0);
    state.testOrderId = data.data.ID;
  });

  test('5.2 订单列表', async ({ request }) => {
    const res = await request.get(`${BASE}/b/orders`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((o: any) => o.ID === state.testOrderId);
    expect(found).toBeTruthy();
    expect(found.status).toBe(0); // 待付款
  });

  test('5.3 订单搜索 (按订单号)', async ({ request }) => {
    // 先获取订单号
    const orderRes = await request.get(`${BASE}/b/orders/${state.testOrderId}`, {
      headers: authHeaders(state.adminToken),
    });
    const orderNo = (await orderRes.json()).data.order_no;

    const res = await request.get(`${BASE}/b/orders?keyword=${orderNo}`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThanOrEqual(1);
    expect(data.data.list[0].order_no).toBe(orderNo);
  });

  test('5.4 订单搜索 (按客户昵称)', async ({ request }) => {
    const res = await request.get(`${BASE}/b/orders?keyword=E2E测试客户`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((o: any) => o.ID === state.testOrderId);
    expect(found).toBeTruthy();
  });

  test('5.5 订单搜索 (按猫咪名)', async ({ request }) => {
    const res = await request.get(`${BASE}/b/orders?keyword=E2E测试猫`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((o: any) => o.ID === state.testOrderId);
    expect(found).toBeTruthy();
  });

  test('5.6 订单搜索 + 状态筛选组合', async ({ request }) => {
    // 搜索 + status=0 (待付款)
    const res = await request.get(`${BASE}/b/orders?keyword=E2E&status=0`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const found = data.data.list.find((o: any) => o.ID === state.testOrderId);
    expect(found).toBeTruthy();

    // 搜索 + status=1 (已完成) - 应该找不到
    const res2 = await request.get(`${BASE}/b/orders?keyword=E2E&status=1`, {
      headers: authHeaders(state.adminToken),
    });
    const data2 = await res2.json();
    expect(data2.code).toBe(0);
    const notFound = (data2.data.list || []).find((o: any) => o.ID === state.testOrderId);
    expect(notFound).toBeFalsy();
  });

  test('5.7 支付订单', async ({ request }) => {
    const res = await request.put(`${BASE}/b/orders/${state.testOrderId}/pay`, {
      headers: authHeaders(state.adminToken),
      data: { pay_method: 'wechat' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    // DB 验证
    const rows = await query('SELECT status, pay_status, pay_method FROM orders WHERE id = ?', [state.testOrderId]);
    expect(rows[0].status).toBe(1);
    expect(rows[0].pay_status).toBe(1);
    expect(rows[0].pay_method).toBe('wechat');
  });

  test('5.8 退款订单', async ({ request }) => {
    const res = await request.put(`${BASE}/b/orders/${state.testOrderId}/refund`, {
      headers: authHeaders(state.adminToken),
      data: { remark: 'E2E测试退款' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const rows = await query('SELECT status, pay_status FROM orders WHERE id = ?', [state.testOrderId]);
    expect(rows[0].status).toBe(3); // 已退款
  });
});

// ============================================================
// 6. 会员卡 & 余额调整 (权限管理)
// ============================================================
test.describe.serial('6. 会员卡 & 余额调整', () => {
  let cardId = 0;

  test('6.1 确保会员卡模板存在', async ({ request }) => {
    await ensureTokens(request);
    const res = await request.get(`${BASE}/b/member-card-templates`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    // 如果没有模板，创建一个
    if (!data.data || data.data.length === 0) {
      const createRes = await request.post(`${BASE}/b/member-card-templates`, {
        headers: authHeaders(state.adminToken),
        data: { name: 'E2E测试卡', min_recharge: 100, discount_rate: 0.9 },
      });
      expect((await createRes.json()).code).toBe(0);
    }
  });

  test('6.2 为测试客户开卡', async ({ request }) => {
    // 清理可能存在的旧卡
    await query('DELETE FROM recharge_records WHERE customer_id = ?', [state.testCustomerId]);
    await query('DELETE FROM member_cards WHERE customer_id = ?', [state.testCustomerId]);
    await query('UPDATE customers SET member_card_id = NULL, member_balance = 0 WHERE id = ?', [state.testCustomerId]);

    const tplRes = await request.get(`${BASE}/b/member-card-templates`, {
      headers: authHeaders(state.adminToken),
    });
    const templates = (await tplRes.json()).data;
    const tpl = templates[0];
    expect(tpl).toBeTruthy();

    const res = await request.post(`${BASE}/b/customers/${state.testCustomerId}/member-card`, {
      headers: authHeaders(state.adminToken),
      data: {
        template_id: tpl.ID,
        recharge_amount: tpl.min_recharge,
        remark: 'E2E测试开卡',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    cardId = data.data.ID;
    expect(data.data.balance).toBe(tpl.min_recharge);

    // DB 验证
    const rows = await query('SELECT balance, status FROM member_cards WHERE id = ?', [cardId]);
    expect(rows[0].status).toBe(1);
    expect(parseFloat(rows[0].balance)).toBe(tpl.min_recharge);

    // 验证客户余额同步
    const custRows = await query('SELECT member_balance, member_card_id FROM customers WHERE id = ?', [state.testCustomerId]);
    expect(custRows[0].member_card_id).toBe(cardId);
  });

  test('6.3 充值', async ({ request }) => {
    const beforeRows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    const beforeBalance = parseFloat(beforeRows[0].balance);

    const res = await request.post(`${BASE}/b/customers/${state.testCustomerId}/recharge`, {
      headers: authHeaders(state.adminToken),
      data: { amount: 200, remark: 'E2E测试充值' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const afterRows = await query('SELECT balance, total_recharge FROM member_cards WHERE id = ?', [cardId]);
    expect(parseFloat(afterRows[0].balance)).toBe(beforeBalance + 200);

    // 验证充值记录
    const records = await query(
      'SELECT type, amount FROM recharge_records WHERE card_id = ? ORDER BY id DESC LIMIT 1', [cardId]
    );
    expect(records[0].type).toBe(1); // 充值
    expect(parseFloat(records[0].amount)).toBe(200);
  });

  test('6.4 管理员调整余额 (增加)', async ({ request }) => {
    const beforeRows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    const beforeBalance = parseFloat(beforeRows[0].balance);

    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.adminToken),
      data: { amount: 50.5, remark: '补偿上次服务问题' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const afterRows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    expect(parseFloat(afterRows[0].balance)).toBeCloseTo(beforeBalance + 50.5, 2);

    // 验证记录类型=4
    const records = await query(
      'SELECT type, amount, remark FROM recharge_records WHERE card_id = ? ORDER BY id DESC LIMIT 1', [cardId]
    );
    expect(records[0].type).toBe(4);
    expect(records[0].remark).toContain('补偿');
  });

  test('6.5 管理员调整余额 (减少)', async ({ request }) => {
    const beforeRows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    const beforeBalance = parseFloat(beforeRows[0].balance);

    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.adminToken),
      data: { amount: -30, remark: '扣除误充金额' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const afterRows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    expect(parseFloat(afterRows[0].balance)).toBeCloseTo(beforeBalance - 30, 2);

    // 验证客户余额同步
    const custRows = await query('SELECT member_balance FROM customers WHERE id = ?', [state.testCustomerId]);
    expect(parseFloat(custRows[0].member_balance)).toBeCloseTo(beforeBalance - 30, 2);
  });

  test('6.6 调整余额不能使余额为负', async ({ request }) => {
    const rows = await query('SELECT balance FROM member_cards WHERE id = ?', [cardId]);
    const balance = parseFloat(rows[0].balance);

    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.adminToken),
      data: { amount: -(balance + 100), remark: '试图超额扣除' },
    });
    const data = await res.json();
    expect(data.code).not.toBe(0); // 应失败
  });

  test('6.7 普通员工不能调整余额 (权限检查)', async ({ request }) => {
    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.staffToken),
      data: { amount: 100, remark: '员工试图调整' },
    });
    expect(res.status()).toBe(403);
  });

  test('6.8 调整余额必须填原因', async ({ request }) => {
    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.adminToken),
      data: { amount: 10 }, // 缺少 remark
    });
    const data = await res.json();
    expect(data.code).not.toBe(0);
  });

  test('6.9 查看充值/消费记录', async ({ request }) => {
    const res = await request.get(`${BASE}/b/customers/${state.testCustomerId}/recharge-records`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.length).toBeGreaterThanOrEqual(3); // 开卡 + 充值 + 调整

    // 验证有 type=4 的调整记录
    const adjustRecords = data.data.filter((r: any) => r.type === 4);
    expect(adjustRecords.length).toBeGreaterThanOrEqual(2);
  });
});

// ============================================================
// 7. 权限管理 (RBAC)
// ============================================================
test.describe.serial('7. 权限管理 (RBAC)', () => {
  test('7.0 确保token', async ({ request }) => { await ensureTokens(request); });

  test('7.1 员工不能创建员工', async ({ request }) => {
    const res = await request.post(`${BASE}/b/staffs`, {
      headers: authHeaders(state.staffToken),
      data: { name: '测试', phone: '19900001111', role: 'staff' },
    });
    expect(res.status()).toBe(403);
  });

  test('7.2 员工不能删除客户', async ({ request }) => {
    const res = await request.delete(`${BASE}/b/customers/${state.testCustomerId}`, {
      headers: authHeaders(state.staffToken),
    });
    expect(res.status()).toBe(403);
  });

  test('7.3 员工不能退款', async ({ request }) => {
    const res = await request.put(`${BASE}/b/orders/${state.testOrderId}/refund`, {
      headers: authHeaders(state.staffToken),
      data: { remark: '员工试图退款' },
    });
    expect(res.status()).toBe(403);
  });

  test('7.4 员工不能取消订单', async ({ request }) => {
    const res = await request.put(`${BASE}/b/orders/${state.testOrderId}/cancel`, {
      headers: authHeaders(state.staffToken),
    });
    expect(res.status()).toBe(403);
  });

  test('7.5 员工不能调整余额', async ({ request }) => {
    const res = await request.put(`${BASE}/b/customers/${state.testCustomerId}/adjust-balance`, {
      headers: authHeaders(state.staffToken),
      data: { amount: 100, remark: '员工试图调整' },
    });
    expect(res.status()).toBe(403);
  });

  test('7.6 员工不能查看营收报表', async ({ request }) => {
    const res = await request.get(`${BASE}/b/dashboard/revenue`, {
      headers: authHeaders(state.staffToken),
    });
    expect(res.status()).toBe(403);
  });

  test('7.7 员工不能查看员工业绩', async ({ request }) => {
    const res = await request.get(`${BASE}/b/dashboard/staff`, {
      headers: authHeaders(state.staffToken),
    });
    expect(res.status()).toBe(403);
  });

  test('7.8 员工可以查看今日概览', async ({ request }) => {
    const res = await request.get(`${BASE}/b/dashboard/overview`, {
      headers: authHeaders(state.staffToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  test('7.9 admin 创建员工时设置密码', async ({ request }) => {
    const testPhone = `198${Date.now().toString().slice(-8)}`;
    const customPwd = 'test8888';
    const res = await request.post(`${BASE}/b/staffs`, {
      headers: authHeaders(state.adminToken),
      data: { name: 'E2E密码测试', phone: testPhone, password: customPwd },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const newStaffId = data.data.ID;

    // 用自定义密码登录
    const loginRes = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: testPhone, password: customPwd },
    });
    const loginData = await loginRes.json();
    expect(loginData.code).toBe(0);
    expect(loginData.data.staff.name).toBe('E2E密码测试');

    // 默认密码不能登录
    const badRes = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: testPhone, password: '123456' },
    });
    expect((await badRes.json()).code).not.toBe(0);

    // admin 重置密码
    const resetRes = await request.put(`${BASE}/b/staffs/${newStaffId}/password`, {
      headers: authHeaders(state.adminToken),
      data: { password: 'newpwd999' },
    });
    expect((await resetRes.json()).code).toBe(0);

    // 新密码可以登录
    const newLoginRes = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: testPhone, password: 'newpwd999' },
    });
    expect((await newLoginRes.json()).code).toBe(0);

    // 旧密码不能登录
    const oldRes = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: testPhone, password: customPwd },
    });
    expect((await oldRes.json()).code).not.toBe(0);

    // 清理
    await query('DELETE FROM staffs WHERE id = ?', [newStaffId]);
  });

  test('7.10 员工不能重置密码', async ({ request }) => {
    const res = await request.put(`${BASE}/b/staffs/2/password`, {
      headers: authHeaders(state.staffToken),
      data: { password: 'hacked123' },
    });
    expect(res.status()).toBe(403);
  });
});

// ============================================================
// 8. 预约管理
// ============================================================
test.describe.serial('8. 预约管理', () => {
  test('8.0 确保token', async ({ request }) => { await ensureTokens(request); });

  test('8.1 获取可用时间段', async ({ request }) => {
    const today = localDate();
    const res = await request.get(`${BASE}/b/appointments/slots?date=${today}&service_id=1`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  test('8.2 创建预约', async ({ request }) => {
    // 用seed数据的客户和宠物避免依赖前面测试
    const customers = await query('SELECT id FROM customers WHERE shop_id = 1 LIMIT 1');
    const pets = await query('SELECT id FROM pets WHERE shop_id = 1 LIMIT 1');
    const custId = state.testCustomerId || customers[0]?.id || 1;
    const petId = state.testPetId || pets[0]?.id || 1;

    const today = localDate();
    // 用唯一时间避免冲突
    const minute = new Date().getMinutes();
    const startTime = `14:${String(minute).padStart(2, '0')}`;
    const res = await request.post(`${BASE}/b/appointments`, {
      headers: authHeaders(state.adminToken),
      data: {
        customer_id: custId,
        pet_id: petId,
        staff_id: 2,
        date: today,
        start_time: startTime,
        service_ids: [1],
        source: 2,
        notes: 'E2E全流程测试预约',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    state.testAppointmentId = data.data.ID;
    expect(data.data.status).toBe(0); // 待确认
  });

  test('8.3 确认预约', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${state.testAppointmentId}/status`, {
      headers: authHeaders(state.adminToken),
      data: { status: 1 },
    });
    expect((await res.json()).code).toBe(0);

    const rows = await query('SELECT status FROM appointments WHERE id = ?', [state.testAppointmentId]);
    expect(rows[0].status).toBe(1);
  });

  test('8.4 开始服务', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${state.testAppointmentId}/status`, {
      headers: authHeaders(state.adminToken),
      data: { status: 2 },
    });
    expect((await res.json()).code).toBe(0);
  });

  test('8.5 完成服务', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${state.testAppointmentId}/status`, {
      headers: authHeaders(state.adminToken),
      data: { status: 3, staff_notes: 'E2E测试-服务完成' },
    });
    expect((await res.json()).code).toBe(0);
  });

  test('8.6 从预约生成订单', async ({ request }) => {
    const res = await request.post(`${BASE}/b/orders/from-appointment`, {
      headers: authHeaders(state.adminToken),
      data: { appointment_id: state.testAppointmentId },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.order_no).toBeTruthy();
    expect(data.data.items.length).toBeGreaterThanOrEqual(1);
  });

  test('8.7 预约列表', async ({ request }) => {
    const today = localDate();
    const res = await request.get(`${BASE}/b/appointments?date=${today}`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });
});

// ============================================================
// 9. 服务管理
// ============================================================
test.describe.serial('9. 服务管理', () => {
  test('9.0 确保token', async ({ request }) => { await ensureTokens(request); });

  test('9.1 服务列表', async ({ request }) => {
    const res = await request.get(`${BASE}/b/services`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThan(0);
  });

  test('9.2 定价查询', async ({ request }) => {
    const res = await request.get(`${BASE}/b/orders/price-lookup?service_id=1&fur_level=短毛猫`, {
      headers: authHeaders(state.adminToken),
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.price).toBeGreaterThan(0);
  });
});

// ============================================================
// 10. 数据一致性 DB 验证
// ============================================================
test.describe.serial('10. 数据一致性验证', () => {
  test('10.1 会员卡余额 = 客户余额', async () => {
    const rows = await query(`
      SELECT c.id, c.member_balance, mc.balance as card_balance
      FROM customers c
      JOIN member_cards mc ON mc.customer_id = c.id AND mc.status = 1
      WHERE c.member_balance != mc.balance
    `);
    expect(rows.length).toBe(0); // 不应有不一致的记录
  });

  test('10.2 充值记录余额连续性', async () => {
    if (!state.testCustomerId) return;
    const records = await query(
      'SELECT type, amount, balance_after FROM recharge_records WHERE customer_id = ? ORDER BY id ASC',
      [state.testCustomerId]
    );
    // 每条记录的 balance_after 应 >= 0
    for (const r of records) {
      expect(parseFloat(r.balance_after)).toBeGreaterThanOrEqual(0);
    }
  });

  test('10.3 订单金额一致性', async () => {
    const rows = await query(`
      SELECT o.id, o.total_amount, COALESCE(SUM(oi.amount), 0) as items_total
      FROM orders o
      LEFT JOIN order_items oi ON oi.order_id = o.id
      GROUP BY o.id
      HAVING ABS(o.total_amount - items_total) > 0.01
    `);
    expect(rows.length).toBe(0);
  });

  test('10.4 宠物的 customer_id 引用有效', async () => {
    const rows = await query(`
      SELECT p.id, p.customer_id
      FROM pets p
      LEFT JOIN customers c ON c.id = p.customer_id
      WHERE p.customer_id IS NOT NULL AND c.id IS NULL
    `);
    expect(rows.length).toBe(0);
  });
});

// ============================================================
// 11. 清理测试数据
// ============================================================
test.describe.serial('11. 清理测试数据', () => {
  test('清理', async () => {
    if (state.testCustomerId) {
      // 先删除有外键依赖的子表
      await query('DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE customer_id = ?)', [state.testCustomerId]);
      await query('DELETE FROM orders WHERE customer_id = ?', [state.testCustomerId]);
      await query('DELETE FROM appointment_services WHERE appointment_id IN (SELECT id FROM appointments WHERE customer_id = ?)', [state.testCustomerId]);
      await query('DELETE FROM appointments WHERE customer_id = ?', [state.testCustomerId]);
      await query('DELETE FROM recharge_records WHERE customer_id = ?', [state.testCustomerId]);
      await query('UPDATE customers SET member_card_id = NULL WHERE id = ?', [state.testCustomerId]);
      await query('DELETE FROM member_cards WHERE customer_id = ?', [state.testCustomerId]);
    }
    if (state.testPetId) await query('DELETE FROM pets WHERE id = ?', [state.testPetId]);
    if (state.testCustomerId) {
      await query('DELETE FROM customers WHERE id = ?', [state.testCustomerId]);
    }
    if (db) await db.end();
  });
});
