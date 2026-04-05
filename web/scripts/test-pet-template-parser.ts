import { parsePetTemplate } from '../src/utils/pet-template-parser'

const dentalSample = `树街の猫|宝贝洁牙小调查

📇基本信息
大名：白菜
品种：橘猫
性别：公
出生日期：2021.7
是否绝育：是
日常饮食习惯：猫粮

🧑🏻‍⚕️健康情况
疫苗是否齐全：是
疾病史和健康情况：嘴巴庞臭

💭性格小秘密
i猫e猫：超级i猫
对陌生环境反应：害怕会躲

请家长阅读洗浴需知’ (诚恳`

const groomingSample = `树街の猫|宝贝洗护小调查

📇基本信息
大名：猪蹄
品种：英短乳白
性别：公
出生日期：2018.8
是否绝育：是
是否毛发打结：否
上次洗澡时间：2024.12

🧑🏻‍⚕️健康情况
疫苗是否齐全：是
疾病史和健康情况：耳朵脏

💭性格小秘密
i猫e猫：e猫
对水/声音/陌生环境反应：轻微抵触

请家长阅读洗浴需知’ (诚恳`

const fuzzyReactionSample = `树街の猫|宝贝洗护小调查

📇基本信息
大名：团子

💭性格小秘密
反应：吹风会挣扎`

function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (actual !== expected) {
    throw new Error(`${label}: expected "${String(expected)}", got "${String(actual)}"`)
  }
}

function runCase(name: string, input: string, expected: Record<string, unknown>) {
  const parsed = parsePetTemplate(input)
  const parsedRecord = parsed as unknown as Record<string, unknown>
  Object.entries(expected).forEach(([key, value]) => {
    assertEqual(parsedRecord[key], value, `${name}.${key}`)
  })
  console.log(`[pass] ${name}`)
  console.log(JSON.stringify(parsed, null, 2))
}

runCase('dental', dentalSample, {
  surveyType: 'dental',
  name: '白菜',
  breed: '橘猫',
  gender: 1,
  birthDate: '2021-7',
  neutered: true,
  dailyDiet: '猫粮',
  furMatted: '',
  lastBathTime: '',
  vaccination: '是',
  healthHistory: '嘴巴庞臭',
  personality: '超级i猫',
  reactions: '害怕会躲',
  reactionsLabel: '对陌生环境反应',
})

runCase('grooming', groomingSample, {
  surveyType: 'grooming',
  name: '猪蹄',
  breed: '英短乳白',
  gender: 1,
  birthDate: '2018-8',
  neutered: true,
  dailyDiet: '',
  furMatted: '否',
  lastBathTime: '2024.12',
  vaccination: '是',
  healthHistory: '耳朵脏',
  personality: 'e猫',
  reactions: '轻微抵触',
  reactionsLabel: '对水/声音/陌生环境反应',
})

runCase('fuzzyReaction', fuzzyReactionSample, {
  surveyType: 'grooming',
  name: '团子',
  reactions: '吹风会挣扎',
  reactionsLabel: '反应',
})

console.log('parser tests passed')
