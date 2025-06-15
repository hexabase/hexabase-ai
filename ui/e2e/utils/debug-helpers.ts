import { Page, TestInfo, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Debug helpers for enhanced E2E testing
 * Provides console error monitoring, step-by-step debugging, and detailed logging
 */

export interface ConsoleEntry {
  type: string;
  text: string;
  timestamp: string;
  location?: any;
  stack?: string;
}

export interface DebugSession {
  sessionId: string;
  sessionDir: string;
  startTime: string;
  testInfo: TestInfo;
}

export class DebugHelper {
  private page: Page;
  private testInfo: TestInfo;
  private session: DebugSession;
  private stepCount = 0;
  private consoleEntries: ConsoleEntry[] = [];
  private errors: ConsoleEntry[] = [];
  private warnings: ConsoleEntry[] = [];

  constructor(page: Page, testInfo: TestInfo) {
    this.page = page;
    this.testInfo = testInfo;
    this.session = this.initializeSession();
    this.setupConsoleMonitoring();
  }

  private initializeSession(): DebugSession {
    const sessionId = process.env.SESSION_ID || new Date().toISOString().replace(/[:.]/g, '-');
    const sessionDir = process.env.SESSION_DIR || path.join(process.cwd(), 'debug-output', sessionId);
    
    // Ensure session directory exists
    if (!fs.existsSync(sessionDir)) {
      fs.mkdirSync(sessionDir, { recursive: true });
    }

    return {
      sessionId,
      sessionDir,
      startTime: new Date().toISOString(),
      testInfo: this.testInfo
    };
  }

  private setupConsoleMonitoring() {
    // Monitor console messages
    this.page.on('console', (msg) => {
      const entry: ConsoleEntry = {
        type: msg.type(),
        text: msg.text(),
        timestamp: new Date().toISOString(),
        location: msg.location()
      };

      this.consoleEntries.push(entry);

      // Log to console with color coding
      const prefix = `[CONSOLE ${entry.type.toUpperCase()}]`;
      console.log(`${prefix} ${entry.text}`);

      switch (msg.type()) {
        case 'error':
          this.errors.push(entry);
          if (process.env.STOP_ON_CONSOLE_ERROR === 'true') {
            console.error('❌ Console error detected - test will be paused');
            this.saveConsoleError(entry);
          }
          break;
        case 'warning':
          this.warnings.push(entry);
          break;
      }
    });

    // Monitor page errors
    this.page.on('pageerror', (err) => {
      const entry: ConsoleEntry = {
        type: 'page-error',
        text: err.message,
        timestamp: new Date().toISOString(),
        stack: err.stack
      };

      this.errors.push(entry);
      console.error('❌ Page error detected:', err.message);
      
      if (process.env.STOP_ON_CONSOLE_ERROR === 'true') {
        this.saveConsoleError(entry);
        throw err;
      }
    });

    // Monitor network failures
    this.page.on('requestfailed', (request) => {
      const entry: ConsoleEntry = {
        type: 'network-error',
        text: `Failed request: ${request.method()} ${request.url()} - ${request.failure()?.errorText}`,
        timestamp: new Date().toISOString()
      };

      this.errors.push(entry);
      console.error('🌐 Network request failed:', entry.text);
    });
  }

  private saveConsoleError(entry: ConsoleEntry) {
    const errorFile = path.join(this.session.sessionDir, 'console', `error-${Date.now()}.json`);
    const consoleDir = path.dirname(errorFile);
    
    if (!fs.existsSync(consoleDir)) {
      fs.mkdirSync(consoleDir, { recursive: true });
    }

    fs.writeFileSync(errorFile, JSON.stringify({
      ...entry,
      testFile: this.testInfo.file,
      testTitle: this.testInfo.title,
      step: this.stepCount,
      url: this.page.url(),
      viewport: this.page.viewportSize()
    }, null, 2));
  }

  /**
   * Execute a step with enhanced debugging
   */
  async step(description: string, action: () => Promise<any>): Promise<any> {
    this.stepCount++;
    const stepNum = this.stepCount.toString().padStart(2, '0');
    
    console.log(`\n🔍 Step ${stepNum}: ${description}`);
    
    // Step-by-step mode handling
    if (process.env.STEP_BY_STEP === 'true') {
      console.log('⏸️  Paused for manual inspection. Press any key to continue...');
      await this.page.waitForTimeout(2000); // In real implementation, would wait for user input
    }

    try {
      // Take screenshot before action
      await this.takeStepScreenshot(`step-${stepNum}-before`);

      // Log page state
      await this.logPageState(`step-${stepNum}-before`);

      // Execute the action
      const result = await action();

      // Take screenshot after action
      await this.takeStepScreenshot(`step-${stepNum}-after`);

      // Log page state after action
      await this.logPageState(`step-${stepNum}-after`);

      console.log(`✅ Step ${stepNum} completed successfully`);
      return result;

    } catch (error) {
      // Take screenshot on error
      await this.takeStepScreenshot(`step-${stepNum}-error`);

      // Log error details
      await this.logError(error as Error, `step-${stepNum}`);

      console.error(`❌ Step ${stepNum} failed: ${(error as Error).message}`);
      throw error;
    }
  }

  /**
   * Take a screenshot with enhanced metadata
   */
  async takeStepScreenshot(name: string): Promise<string> {
    const screenshotPath = path.join(this.session.sessionDir, 'screenshots', `${name}.png`);
    const screenshotDir = path.dirname(screenshotPath);
    
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }

    await this.page.screenshot({
      path: screenshotPath,
      fullPage: true,
      animations: 'disabled'
    });

    return screenshotPath;
  }

  /**
   * Log current page state
   */
  async logPageState(stepName: string) {
    const stateFile = path.join(this.session.sessionDir, 'logs', `${stepName}-state.json`);
    const logsDir = path.dirname(stateFile);
    
    if (!fs.existsSync(logsDir)) {
      fs.mkdirSync(logsDir, { recursive: true });
    }

    const state = {
      timestamp: new Date().toISOString(),
      url: this.page.url(),
      title: await this.page.title(),
      viewport: this.page.viewportSize(),
      cookies: await this.page.context().cookies(),
      localStorage: await this.page.evaluate(() => ({ ...localStorage })),
      sessionStorage: await this.page.evaluate(() => ({ ...sessionStorage })),
      visibleElements: await this.getVisibleElements(),
      consoleErrors: this.errors.length,
      consoleWarnings: this.warnings.length
    };

    fs.writeFileSync(stateFile, JSON.stringify(state, null, 2));
  }

  /**
   * Get visible elements for debugging
   */
  private async getVisibleElements(): Promise<string[]> {
    return await this.page.evaluate(() => {
      const elements: string[] = [];
      const visibleElements = document.querySelectorAll('*');
      
      visibleElements.forEach((el) => {
        const rect = el.getBoundingClientRect();
        if (rect.width > 0 && rect.height > 0 && 
            rect.top >= 0 && rect.left >= 0 &&
            rect.bottom <= window.innerHeight && 
            rect.right <= window.innerWidth) {
          const tag = el.tagName.toLowerCase();
          const id = el.id ? `#${el.id}` : '';
          const classes = el.className ? `.${el.className.split(' ').join('.')}` : '';
          elements.push(`${tag}${id}${classes}`);
        }
      });
      
      return elements.slice(0, 50); // Limit to first 50 visible elements
    });
  }

  /**
   * Log error details
   */
  async logError(error: Error, stepName: string) {
    const errorFile = path.join(this.session.sessionDir, 'logs', `${stepName}-error.json`);
    const logsDir = path.dirname(errorFile);
    
    if (!fs.existsSync(logsDir)) {
      fs.mkdirSync(logsDir, { recursive: true });
    }

    const errorDetails = {
      timestamp: new Date().toISOString(),
      step: stepName,
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack
      },
      page: {
        url: this.page.url(),
        title: await this.page.title(),
        viewport: this.page.viewportSize()
      },
      console: {
        errors: this.errors,
        warnings: this.warnings,
        totalEntries: this.consoleEntries.length
      },
      test: {
        file: this.testInfo.file,
        title: this.testInfo.title,
        timeout: this.testInfo.timeout
      }
    };

    fs.writeFileSync(errorFile, JSON.stringify(errorDetails, null, 2));
  }

  /**
   * Wait with enhanced debugging
   */
  async waitForSelector(selector: string, options?: { timeout?: number; state?: 'attached' | 'detached' | 'visible' | 'hidden' }): Promise<void> {
    const timeout = options?.timeout || 30000;
    const state = options?.state || 'visible';
    
    console.log(`⏳ Waiting for selector: ${selector} (state: ${state}, timeout: ${timeout}ms)`);
    
    try {
      await expect(this.page.locator(selector)).toHaveState(state as any, { timeout });
      console.log(`✅ Element found: ${selector}`);
    } catch (error) {
      console.error(`❌ Element not found: ${selector}`);
      await this.takeStepScreenshot(`wait-failed-${selector.replace(/[^a-zA-Z0-9]/g, '-')}`);
      throw error;
    }
  }

  /**
   * Click with enhanced debugging
   */
  async click(selector: string, options?: { timeout?: number }): Promise<void> {
    return this.step(`Click on ${selector}`, async () => {
      await this.waitForSelector(selector);
      await this.page.locator(selector).click(options);
    });
  }

  /**
   * Fill input with enhanced debugging
   */
  async fill(selector: string, value: string, options?: { timeout?: number }): Promise<void> {
    return this.step(`Fill "${selector}" with "${value}"`, async () => {
      await this.waitForSelector(selector);
      await this.page.locator(selector).fill(value, options);
    });
  }

  /**
   * Navigate with enhanced debugging
   */
  async goto(url: string, options?: { timeout?: number; waitUntil?: 'load' | 'domcontentloaded' | 'networkidle' }): Promise<void> {
    return this.step(`Navigate to ${url}`, async () => {
      await this.page.goto(url, options);
    });
  }

  /**
   * Generate debug report
   */
  generateDebugReport(): object {
    const report = {
      session: this.session,
      test: {
        file: this.testInfo.file,
        title: this.testInfo.title,
        duration: Date.now() - new Date(this.session.startTime).getTime(),
        steps: this.stepCount
      },
      console: {
        totalEntries: this.consoleEntries.length,
        errors: this.errors.length,
        warnings: this.warnings.length,
        details: {
          errors: this.errors,
          warnings: this.warnings
        }
      },
      browser: {
        url: this.page.url(),
        viewport: this.page.viewportSize()
      }
    };

    // Save report to file
    const reportFile = path.join(this.session.sessionDir, 'debug-report.json');
    fs.writeFileSync(reportFile, JSON.stringify(report, null, 2));

    return report;
  }

  /**
   * Check for console errors
   */
  hasConsoleErrors(): boolean {
    return this.errors.length > 0;
  }

  /**
   * Get console error summary
   */
  getConsoleErrorSummary(): { count: number; errors: ConsoleEntry[] } {
    return {
      count: this.errors.length,
      errors: this.errors
    };
  }
}