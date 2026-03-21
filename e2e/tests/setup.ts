// Script to seed initial test data directly into MySQL
// Run: npx tsx tests/setup.ts

async function setup() {
  const BASE = 'http://localhost:8080/api/v1';

  // Health check
  const health = await fetch(`${BASE}/health`);
  const hData = await health.json();
  console.log('Health:', hData);

  // Try login - if it fails, the seed data doesn't exist yet
  // We need to create it via a separate Go seed script
  console.log('Server is running. Test data needs to be seeded via Go.');
  console.log('Run: cd server && go run cmd/seed/main.go');
}

setup().catch(console.error);
