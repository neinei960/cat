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
  open_time: string
  close_time: string
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
  sort_order?: number
  status: number
  commission_rate: number
  product_commission_rate?: number
  feeding_commission_rate?: number
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
  address: string
  address_detail: string
  door_code: string
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
  card_type: number // 1储值卡 2次卡
  min_recharge: number
  discount_rate: number
  product_discount_rate: number
  valid_days: number
  total_times: number // 次卡总次数
  times_price: number // 次卡售价
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
  product_discount_rate: number
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

interface PetBathReport {
  ID: number
  pet_id: number
  shop_id: number
  image_url: string
  CreatedAt: string
}

interface ProductSKU {
  ID: number
  product_id: number
  spec_name: string
  price: number
  weight: number
  sellable: boolean
}

interface Product {
  ID: number
  shop_id: number
  category_id: number
  name: string
  brand: string
  description: string
  multi_spec: boolean
  status: number
  category?: {
    ID: number
    name: string
  }
  skus?: ProductSKU[]
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
  pricing_type: number  // 1按次 2按天
  holiday_price: number
  monthly_usage_count?: number
  price_rules?: ServicePriceRule[]
  discounts?: ServiceDiscount[]
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

interface ServiceDiscount {
  ID: number
  service_id: number
  type: number        // 1满天折扣 2住N免M
  min_days: number
  discount_price: number
  free_days: number
  is_holiday: boolean
  status: number
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
  staff?: Staff
}

interface BoardingCabinet {
  ID: number
  shop_id: number
  code: string
  cabinet_type: string
  room_count: number
  capacity: number
  base_price: number
  extra_pet_price: number
  status: string
  remark: string
  occupied_rooms?: number
  reserved_rooms?: number
  remaining_rooms?: number
  CreatedAt?: string
}

interface BoardingHoliday {
  ID: number
  shop_id: number
  holiday_date: string
  name: string
  CreatedAt?: string
}

interface BoardingDiscountPolicy {
  ID: number
  shop_id: number
  name: string
  policy_type: string
  rule_json: string
  valid_from: string
  valid_to: string
  priority: number
  stackable: boolean
  status: number
  remark: string
  rule?: any
}

interface BoardingPriceLine {
  type: string
  label: string
  quantity: number
  unit_price: number
  amount: number
}

interface BoardingPricePreview {
  check_in_at: string
  check_out_at: string
  nights: number
  pet_count: number
  regular_nights: number
  holiday_nights: number
  base_amount: number
  extra_pet_amount: number
  holiday_surcharge_amount: number
  discount_amount: number
  pay_amount: number
  policies: BoardingDiscountPolicy[]
  lines: BoardingPriceLine[]
}

interface BoardingOrderPet {
  ID: number
  boarding_order_id: number
  pet_id: number
  pet_name_snapshot: string
  remark: string
  pet?: Pet
}

interface BoardingOrderLog {
  ID: number
  boarding_order_id: number
  operator_id: number
  action: string
  content: string
  operator?: Staff
  CreatedAt?: string
}

interface BoardingOrder {
  ID: number
  shop_id: number
  order_id?: number
  customer_id: number
  staff_id: number
  cabinet_id: number
  check_in_at: string
  check_out_at: string
  actual_check_out_at: string
  nights: number
  base_amount: number
  holiday_surcharge_amount: number
  discount_amount: number
  manual_discount_amount: number
  pay_amount: number
  status: string
  has_deworming?: boolean | null
  remark: string
  policy_snapshot_json: string
  price_snapshot_json: string
  customer?: Customer
  staff?: Staff
  cabinet?: BoardingCabinet
  pets?: BoardingOrderPet[]
  logs?: BoardingOrderLog[]
  order?: Order
  CreatedAt?: string
}

interface BoardingDashboardGroup {
  cabinet_id: number
  cabinet_type: string
  room_count: number
  capacity: number
  base_price: number
  extra_pet_price: number
  status: string
  remark: string
  occupied_rooms: number
  reserved_rooms: number
  remaining_rooms: number
  orders: BoardingOrder[]
}

interface FeedingPricingSetting {
  base_day_price: number
  holiday_day_price: number
  discount_day_price: number
  discount_holiday_price: number
  discount_start_day: number
}

interface FeedingItemTemplate {
  code: string
  name: string
  extra_price: number
}

interface FeedingSettings {
  pricing: FeedingPricingSetting
  items: FeedingItemTemplate[]
}

interface FeedingAddressSnapshot {
  address: string
  detail?: string
  door_code?: string
}

interface FeedingPlanPet {
  ID?: number
  feeding_plan_id?: number
  pet_id: number
  pet_name_snapshot?: string
  remark?: string
  pet?: Pet
}

interface FeedingPlanRule {
  ID?: number
  feeding_plan_id?: number
  weekday: number
  window_code: string
  visit_count: number
}

interface FeedingVisitItem {
  ID: number
  feeding_visit_id: number
  item_code: string
  item_name_snapshot: string
  extra_price: number
  checked: boolean
}

interface FeedingVisitLog {
  ID: number
  feeding_visit_id: number
  operator_id: number
  action: string
  content: string
  operator?: Staff
  CreatedAt?: string
}

interface FeedingVisitMedia {
  ID: number
  feeding_visit_id: number
  media_type: string
  url: string
  CreatedAt?: string
}

interface FeedingVisit {
  ID: number
  shop_id: number
  feeding_plan_id: number
  scheduled_date: string
  window_code: string
  staff_id?: number
  status: string
  visit_price: number
  arrived_at?: string
  completed_at?: string
  customer_note?: string
  internal_note?: string
  exception_type?: string
  plan?: FeedingPlan
  staff?: Staff
  items?: FeedingVisitItem[]
  logs?: FeedingVisitLog[]
  media?: FeedingVisitMedia[]
  CreatedAt?: string
}

interface FeedingPlan {
  ID: number
  shop_id: number
  customer_id: number
  order_id?: number
  address_snapshot_json: string
  contact_name: string
  contact_phone: string
  start_date: string
  end_date: string
  time_granularity: string
  status: string
  remark: string
  pricing_snapshot_json: string
  selected_items_json: string
  selected_dates_json: string
  play_dates_json: string
  play_mode: string
  play_count: number
  other_price: number
  deposit: number
  total_amount: number
  unpaid_amount: number
  customer?: Customer
  order?: Order
  pets?: FeedingPlanPet[]
  rules?: FeedingPlanRule[]
  visits?: FeedingVisit[]
  CreatedAt?: string
}

interface FeedingDashboardGroup {
  status: string
  label: string
  count: number
  visits: FeedingVisit[]
}

interface FeedingDashboardResponse {
  date: string
  groups: FeedingDashboardGroup[]
}

interface Order {
  ID: number
  order_no: string
  customer_id?: number
  pet_id?: number
  appointment_id?: number
  feeding_plan_id?: number
  staff_id?: number
  total_amount: number
  service_total?: number
  product_total?: number
  addon_total?: number
  discount_amount: number
  service_discount_amount?: number
  product_discount_amount?: number
  discount_rate: number
  pay_amount: number
  commission: number
  pay_method: string
  pay_status: number
  status: number
  remark: string
  order_kind?: 'service' | 'product' | 'mixed' | 'feeding'
  pet_summary?: string
  pet_groups?: OrderPetGroup[]
  customer?: Customer
  pet?: Pet
  staff?: Staff
  items?: OrderItem[]
  pay_time?: string
  CreatedAt?: string
}

interface OrderItem {
  ID: number
  order_id: number
  item_type: number
  item_id: number
  name: string
  quantity: number
  unit_price: number
  amount: number
}

interface OrderPetGroup {
  pet_id?: number
  pet_name: string
  items: OrderItem[]
}
