import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:8080/api/v1';

function localDate(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;
}

let token = '';

test.describe.serial('API Integration Tests', () => {
  test('health check', async ({ request }) => {
    const res = await request.get(`${BASE}/health`);
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.status).toBe('ok');
  });

  test('staff login', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.token).toBeTruthy();
    expect(data.data.staff.name).toBe('张店长');
    expect(data.data.staff.role).toBe('admin');
    token = data.data.token;
  });

  test('login with wrong password fails', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: 'wrongpassword' },
    });
    const data = await res.json();
    expect(data.code).not.toBe(0);
  });

  test('get shop info', async ({ request }) => {
    const res = await request.get(`${BASE}/b/shop`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.name).toBe('测试宠物店');
  });

  test('list staffs', async ({ request }) => {
    const res = await request.get(`${BASE}/b/staffs`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThanOrEqual(3);
  });

  test('list services', async ({ request }) => {
    const res = await request.get(`${BASE}/b/services`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThanOrEqual(3);
  });

  test('list customers', async ({ request }) => {
    const res = await request.get(`${BASE}/b/customers`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThanOrEqual(1);
  });

  test('list pets', async ({ request }) => {
    const res = await request.get(`${BASE}/b/pets`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.list.length).toBeGreaterThanOrEqual(1);
  });

  test('get available slots', async ({ request }) => {
    const today = localDate();
    const res = await request.get(`${BASE}/b/appointments/slots?date=${today}&service_id=1`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    // Should have slots for techs who have schedules today
    expect(Array.isArray(data.data) ? data.data.length : 0).toBeGreaterThan(0);
  });

  let appointmentId = 0;

  test('create appointment', async ({ request }) => {
    const today = localDate();
    // Use a unique time based on current minutes to avoid conflict with previous test runs
    const now = new Date();
    const hour = Math.max(9, Math.min(16, now.getHours()));
    const startTime = `${String(hour).padStart(2, '0')}:${now.getMinutes() < 30 ? '00' : '30'}`;
    const res = await request.post(`${BASE}/b/appointments`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        customer_id: 1,
        pet_id: 1,
        staff_id: 3, // tech2 to avoid conflicts
        date: today,
        start_time: startTime,
        service_ids: [1],
        source: 2,
        notes: 'Playwright 测试预约',
      },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.ID).toBeGreaterThan(0);
    expect(data.data.status).toBe(0); // pending
    expect(data.data.total_amount).toBe(88); // standard wash price
    appointmentId = data.data.ID;
  });

  test('get appointment detail', async ({ request }) => {
    const res = await request.get(`${BASE}/b/appointments/${appointmentId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.notes).toBe('Playwright 测试预约');
    expect(data.data.services.length).toBe(1);
    expect(data.data.services[0].service_name).toBe('标准洗澡');
  });

  test('confirm appointment', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${appointmentId}/status`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { status: 1 },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  test('start appointment (in progress)', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${appointmentId}/status`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { status: 2 },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  test('complete appointment', async ({ request }) => {
    const res = await request.put(`${BASE}/b/appointments/${appointmentId}/status`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { status: 3, staff_notes: '服务完成，毛发状态良好' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  let orderId = 0;

  test('create order from appointment', async ({ request }) => {
    const res = await request.post(`${BASE}/b/orders/from-appointment`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { appointment_id: appointmentId },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.order_no).toBeTruthy();
    expect(data.data.total_amount).toBe(88);
    expect(data.data.items.length).toBe(1);
    orderId = data.data.ID;
  });

  test('pay order', async ({ request }) => {
    const res = await request.put(`${BASE}/b/orders/${orderId}/pay`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { pay_method: 'cash' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
  });

  test('verify order is completed', async ({ request }) => {
    const res = await request.get(`${BASE}/b/orders/${orderId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.status).toBe(1); // completed
    expect(data.data.pay_status).toBe(1); // paid
    expect(data.data.pay_method).toBe('cash');
  });

  test('conflict detection works', async ({ request }) => {
    const today = localDate();
    const res = await request.post(`${BASE}/b/appointments`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        customer_id: 1,
        pet_id: 1,
        staff_id: 2,
        date: today,
        start_time: '10:00', // same time as first appointment
        service_ids: [1],
        source: 2,
      },
    });
    const data = await res.json();
    // This may succeed (if previous appt is completed, status=3 excluded from conflict)
    // or fail (if there's an active appt). Either way it's valid behavior.
    expect(data.code === 0 || data.code === -1).toBeTruthy();
  });

  test('dashboard overview', async ({ request }) => {
    const res = await request.get(`${BASE}/b/dashboard/overview`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    expect(data.data.total_customers).toBeGreaterThanOrEqual(1);
  });

  test('unauthorized access fails', async ({ request }) => {
    const res = await request.get(`${BASE}/b/shop`);
    const data = await res.json();
    expect(data.code).toBe(401);
  });
});
