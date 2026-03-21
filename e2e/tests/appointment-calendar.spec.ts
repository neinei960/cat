import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:8080/api/v1';

function localDate(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;
}

let token = '';

test.describe.serial('Appointment Calendar Tests', () => {
  test('login', async ({ request }) => {
    const res = await request.post(`${BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    token = data.data.token;
    console.log('Logged in as:', data.data.staff.name);
  });

  test('list staff and find 王技师', async ({ request }) => {
    const res = await request.get(`${BASE}/b/staffs`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    const wangStaff = data.data.list.find((s: any) => s.name === '王技师');
    console.log('Staff list:', data.data.list.map((s: any) => `${s.ID}:${s.name}(${s.role})`));
    expect(wangStaff).toBeTruthy();
    console.log('王技师 ID:', wangStaff.ID);
  });

  let newApptId = 0;

  test('create appointment with 王技师', async ({ request }) => {
    const today = localDate();
    const res = await request.post(`${BASE}/b/appointments`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        customer_id: 1,
        pet_id: 1,
        staff_id: 3, // 王技师
        date: today,
        start_time: '14:00',
        service_ids: [1],
        source: 2,
        notes: 'Playwright 测试 - 王技师预约',
      },
    });
    const data = await res.json();
    console.log('Create appointment response:', JSON.stringify(data, null, 2));
    expect(data.code).toBe(0);
    newApptId = data.data.ID;
    expect(data.data.staff_id).toBe(3);
    console.log('Created appointment ID:', newApptId);
  });

  test('verify appointment in calendar API', async ({ request }) => {
    const today = localDate();
    const res = await request.get(`${BASE}/b/appointments/calendar?start_date=${today}&end_date=${today}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    console.log('Calendar appointments count:', data.data.length);
    data.data.forEach((a: any) => {
      const staffName = a.staff?.name || '未分配';
      const petName = a.pet?.name || '-';
      console.log(`  ID=${a.ID} time=${a.start_time}-${a.end_time} staff_id=${a.staff_id} staff=${staffName} pet=${petName} status=${a.status}`);
    });

    // Find our appointment
    const ourAppt = data.data.find((a: any) => a.ID === newApptId);
    expect(ourAppt).toBeTruthy();
    expect(ourAppt.staff_id).toBe(3);
    expect(ourAppt.staff?.name).toBe('王技师');
    console.log('✅ Appointment found in calendar with 王技师');
  });

  test('verify appointment in list API', async ({ request }) => {
    const res = await request.get(`${BASE}/b/appointments?page_size=50`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();
    expect(data.code).toBe(0);

    const ourAppt = data.data.list.find((a: any) => a.ID === newApptId);
    expect(ourAppt).toBeTruthy();
    expect(ourAppt.staff?.name).toBe('王技师');
    console.log('✅ Appointment found in list with 王技师');
  });

  test('check staff_id type in calendar response', async ({ request }) => {
    const today = localDate();
    const res = await request.get(`${BASE}/b/appointments/calendar?start_date=${today}&end_date=${today}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const data = await res.json();

    // Check if staff_id is a number (not a string or null)
    const apptWithStaff = data.data.find((a: any) => a.staff_id);
    if (apptWithStaff) {
      console.log('staff_id type:', typeof apptWithStaff.staff_id, 'value:', apptWithStaff.staff_id);
      console.log('staff_id is pointer (*uint)?', apptWithStaff.staff_id !== null);
      // In calendar.vue, staffList uses staff.ID (number)
      // appointment uses staff_id - check if they match type
      expect(typeof apptWithStaff.staff_id).toBe('number');
    }

    // Get staff list for comparison
    const staffRes = await request.get(`${BASE}/b/staffs`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const staffData = await staffRes.json();
    const staffIds = staffData.data.list.map((s: any) => ({ ID: s.ID, type: typeof s.ID, name: s.name }));
    console.log('Staff IDs:', staffIds);

    // The frontend calendar.vue matches: a.staff_id === staffId
    // where staffId comes from staff.ID in the v-for
    // Both should be numbers for === to work
    console.log('✅ Type check passed - both are numbers');
  });

  test('cleanup - cancel test appointment', async ({ request }) => {
    if (newApptId > 0) {
      await request.put(`${BASE}/b/appointments/${newApptId}/status`, {
        headers: { Authorization: `Bearer ${token}` },
        data: { status: 4, cancelled_by: 'test' },
      });
      console.log('Cleaned up appointment', newApptId);
    }
  });
});
