import { FullConfig } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Global setup function for E2E tests
 * Runs once before all tests
 */
async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting E2E test global setup...');
  
  // Create necessary directories
  const directories = [
    'test-results',
    'test-results/screenshots',
    'test-results/videos',
    'test-results/traces',
    'test-results/downloads',
    'debug-output',
  ];
  
  for (const dir of directories) {
    const dirPath = path.join(config.rootDir, dir);
    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath, { recursive: true });
      console.log(`📁 Created directory: ${dir}`);
    }
  }
  
  // Set environment variables for debugging
  process.env.NODE_ENV = 'test';
  process.env.DEBUG = process.env.DEBUG || 'pw:api';
  
  // Log test configuration
  console.log('📋 Test Configuration:');
  console.log(`  - Base URL: ${config.use?.baseURL || 'http://localhost:3000'}`);
  console.log(`  - Headless: ${config.use?.headless !== false}`);
  console.log(`  - Workers: ${config.workers || 1}`);
  console.log(`  - Timeout: ${config.use?.actionTimeout || 30000}ms`);
  
  // Check if services are running
  try {
    const fetch = (await import('node-fetch')).default;
    
    // Check API
    const apiUrl = process.env.API_URL || 'http://localhost:8080';
    try {
      const apiResponse = await fetch(`${apiUrl}/health`, { 
        timeout: 5000,
        signal: AbortSignal.timeout(5000)
      });
      if (apiResponse.ok) {
        console.log('✅ API is running at', apiUrl);
      } else {
        console.warn('⚠️  API returned status:', apiResponse.status);
      }
    } catch (error) {
      console.warn('⚠️  API is not accessible at', apiUrl);
      console.warn('   Make sure to run: make debug-api');
    }
    
    // Check UI
    const uiUrl = config.use?.baseURL || 'http://localhost:3000';
    try {
      const uiResponse = await fetch(uiUrl, { 
        timeout: 5000,
        signal: AbortSignal.timeout(5000)
      });
      if (uiResponse.ok) {
        console.log('✅ UI is running at', uiUrl);
      } else {
        console.warn('⚠️  UI returned status:', uiResponse.status);
      }
    } catch (error) {
      console.warn('⚠️  UI is not accessible at', uiUrl);
      console.warn('   Make sure to run: make debug-ui');
    }
  } catch (error) {
    console.error('❌ Error checking services:', error);
  }
  
  // Create test session info
  const sessionInfo = {
    startTime: new Date().toISOString(),
    config: {
      baseURL: config.use?.baseURL,
      headless: config.use?.headless !== false,
      workers: config.workers || 1,
      projects: config.projects?.map(p => p.name) || [],
    },
    environment: {
      NODE_ENV: process.env.NODE_ENV,
      DEBUG: process.env.DEBUG,
      CI: process.env.CI || false,
    },
  };
  
  fs.writeFileSync(
    path.join(config.rootDir, 'test-results', 'session-info.json'),
    JSON.stringify(sessionInfo, null, 2)
  );
  
  console.log('✅ Global setup completed\n');
}

export default globalSetup;