// Test script to verify development authentication flow
const axios = require('axios');

const API_URL = 'http://localhost:8080';
const DEV_TOKEN = 'dev_token_dev-user-1_1234567890';

async function testAuth() {
  console.log('Testing development authentication...\n');

  try {
    // Test 1: Check if auth endpoint rejects without token
    console.log('1. Testing without token:');
    try {
      await axios.get(`${API_URL}/api/v1/organizations/`);
      console.log('❌ FAIL: Request should have been rejected');
    } catch (error) {
      if (error.response?.status === 401) {
        console.log('✅ PASS: Correctly rejected with 401');
      } else {
        console.log('❌ FAIL: Unexpected error:', error.message);
      }
    }

    // Test 2: Check if auth endpoint accepts dev token
    console.log('\n2. Testing with development token:');
    try {
      const response = await axios.get(`${API_URL}/api/v1/organizations/`, {
        headers: {
          'Authorization': `Bearer ${DEV_TOKEN}`
        }
      });
      console.log('✅ PASS: Request accepted');
      console.log('Response:', JSON.stringify(response.data, null, 2));
    } catch (error) {
      console.log('❌ FAIL: Request rejected');
      console.log('Status:', error.response?.status);
      console.log('Error:', JSON.stringify(error.response?.data || error.message));
    }

    // Test 3: Check if we get the development organization
    console.log('\n3. Verifying development organization:');
    try {
      const response = await axios.get(`${API_URL}/api/v1/organizations/`, {
        headers: {
          'Authorization': `Bearer ${DEV_TOKEN}`
        }
      });
      
      const orgs = response.data.organizations || [];
      const devOrg = orgs.find(org => org.id === 'dev-org-1');
      
      if (devOrg) {
        console.log('✅ PASS: Development organization found');
        console.log('Organization:', JSON.stringify(devOrg, null, 2));
      } else {
        console.log('❌ FAIL: Development organization not found');
      }
    } catch (error) {
      console.log('❌ FAIL: Could not fetch organizations');
      console.log('Error:', error.response?.data || error.message);
    }

  } catch (error) {
    console.error('Test failed with error:', error.message);
  }
}

// Run the test
testAuth();