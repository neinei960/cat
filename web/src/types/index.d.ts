// Common
interface PageParams {
  page?: number
  page_size?: number
}

interface PageResult<T> {
  list: T[]
  total: number
}

// Shop
interface Shop {
  ID: number
  name: string
  logo: string
  phone: string
  address: string
  latitude: number
  longitude: number
  business_hours: any
  status: number
}

// Staff
interface Staff {
  ID: number
  shop_id: number
  phone: string
  name: string
  avatar: string
  role: string
  status: number
  commission_rate: number
  last_login_at: string
}

interface CreateStaffReq {
  phone: string
  name: string
  password?: string
  role?: string
  commission_rate?: number
  product_commission_rate?: number
  feeding_commission_rate?: number
}

// Customer
interface Customer {
  ID: number
  shop_id: number
  openid: string
  phone: string
  nickname: string
  avatar: string
  gender: number
  remark: string
  tags: string
  total_spent: number
  visit_count: number
  last_visit_at: string
  member_balance: number
  discount_rate: number
  member_card_id?: number
  member_card?: MemberCard
  customer_tags?: CustomerTag[]
  pets?: Pet[]
  CreatedAt: string
}

interface CustomerTag {
  ID: number
  shop_id: number
  name: string
  description: string
  color: string
  sort_order: number
  status: number
  relation_count?: number
  CreatedAt?: string
}

// Member Card
interface MemberCardTemplate {
  ID: number
  shop_id: number
  name: string
  min_recharge: number
  discount_rate: number
  valid_days: number
  color: string
  status: number
  sort_order: number
  discounts?: MemberCardDiscount[]
}

interface MemberCardDiscount {
  ID: number
  template_id: number
  category_id: number
  category_name: string
  discount_rate: number
}

interface MemberCard {
  ID: number
  shop_id: number
  customer_id: number
  template_id: number
  card_name: string
  balance: number
  total_recharge: number
  total_spent: number
  discount_rate: number
  expire_at: string | null
  status: number
  template?: MemberCardTemplate
}

interface RechargeRecord {
  ID: number
  shop_id: number
  customer_id: number
  card_id: number
  type: number  // 1充值 2消费 3退款
  amount: number
  balance_after: number
  order_id?: number
  remark: string
  CreatedAt: string
}

// Pet
interface Pet {
  ID: number
  shop_id: number
  customer_id: number
  name: string
  species: string
  breed: string
  gender: number
  birth_date: string
  weight: number
  coat_type: string
  coat_color: string
  fur_level: string
  personality: string
  aggression: string
  forbidden_zones: string
  bath_frequency: string
  neutered: boolean
  avatar: string
  care_notes: string
  behavior_notes: string
  status: number
  customer?: Customer
  CreatedAt: string
}

// Service Category
interface ServiceCategory {
  ID: number
  shop_id: number
  parent_id: number | null
  name: string
  sort_order: number
  status: number
  children?: ServiceCategory[]
}

// Service
interface ServiceItem {
  ID: number
  shop_id: number
  name: string
  category: string
  category_id?: number
  description: string
  base_price: number
  duration: number
  applicable_species: string
  applicable_sizes: string
  sort_order: number
  status: number
  price_rules?: ServicePriceRule[]
  service_category?: ServiceCategory
}

interface ServicePriceRule {
  ID: number
  service_id: number
  name: string
  fur_level: string
  pet_size: string
  breed: string
  price: number
  duration: number
}

// Schedule
interface StaffSchedule {
  ID: number
  staff_id: number
  date: string
  start_time: string
  end_time: string
  break_start: string
  break_end: string
  max_capacity: number
  is_day_off: boolean
}
