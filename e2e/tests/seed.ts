// Seed test data via API
const BASE = 'http://localhost:8080/api/v1';

export async function seedTestData() {
  // Create a shop first via direct DB — for now we'll create staff login
  // The staff login test will create staff via direct API after we have a token

  // For MVP testing, we need to:
  // 1. Have a shop in DB (ID=1)
  // 2. Have a staff with known phone/password
  // We'll use a setup script that inserts via Go
}

export async function loginAsStaff(): Promise<string> {
  const res = await fetch(`${BASE}/auth/staff/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone: '13800138000', password: '123456' }),
  });
  const data = await res.json();
  if (data.code !== 0) throw new Error('Login failed: ' + data.msg);
  return data.data.token;
}

export async function apiRequest(token: string, method: string, path: string, body?: any) {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  return res.json();
}
