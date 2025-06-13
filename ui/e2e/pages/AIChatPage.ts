import { Page, Locator } from '@playwright/test';

export class AIChatPage {
  readonly page: Page;
  readonly chatWindow: Locator;
  readonly chatInput: Locator;
  readonly sendButton: Locator;
  readonly messagesContainer: Locator;
  readonly typingIndicator: Locator;
  readonly settingsButton: Locator;
  readonly historyButton: Locator;
  readonly clearButton: Locator;
  readonly fileInput: Locator;
  readonly commandMenu: Locator;

  constructor(page: Page) {
    this.page = page;
    this.chatWindow = page.getByTestId('ai-chat-window');
    this.chatInput = page.getByTestId('chat-input');
    this.sendButton = page.getByTestId('send-message-button');
    this.messagesContainer = page.getByTestId('chat-messages');
    this.typingIndicator = page.getByTestId('ai-typing-indicator');
    this.settingsButton = page.getByTestId('ai-chat-settings-button');
    this.historyButton = page.getByTestId('chat-history-button');
    this.clearButton = page.getByTestId('clear-conversation-button');
    this.fileInput = page.getByTestId('chat-file-input');
    this.commandMenu = page.getByTestId('ai-command-menu');
  }

  async open() {
    const assistantButton = this.page.getByTestId('ai-assistant-button');
    await assistantButton.click();
    await this.chatWindow.waitFor({ state: 'visible' });
  }

  async close() {
    await this.page.keyboard.press('Escape');
    await this.chatWindow.waitFor({ state: 'hidden' });
  }

  async sendMessage(message: string) {
    await this.chatInput.fill(message);
    await this.sendButton.click();
  }

  async waitForResponse() {
    // Wait for typing indicator to appear and disappear
    await this.typingIndicator.waitFor({ state: 'visible' });
    await this.typingIndicator.waitFor({ state: 'hidden', timeout: 30000 });
  }

  async getLastMessage(type: 'user' | 'ai'): Promise<Locator> {
    return this.page.locator(`[data-testid="chat-message-${type}"]`).last();
  }

  async getMessageCount(): Promise<number> {
    return this.page.locator('[data-testid^="chat-message-"]').count();
  }

  async selectModel(model: string) {
    await this.settingsButton.click();
    const dialog = this.page.getByRole('dialog');
    await dialog.getByTestId('ai-model-select').selectOption(model);
    await dialog.getByTestId('save-settings-button').click();
  }

  async clearConversation() {
    await this.clearButton.click();
    const confirmDialog = this.page.getByRole('dialog');
    await confirmDialog.getByTestId('confirm-clear-button').click();
  }

  async uploadFile(filePath: string | { name: string; mimeType: string; buffer: Buffer }) {
    await this.fileInput.setInputFiles(filePath);
  }

  async useSlashCommand(command: string) {
    await this.chatInput.fill('/');
    await this.commandMenu.waitFor({ state: 'visible' });
    await this.commandMenu.getByTestId(`command-${command}`).click();
  }

  async executeAction(actionType: string) {
    const actionButton = this.page.getByTestId(`ai-action-${actionType}`);
    await actionButton.click();
  }

  async openHistory() {
    await this.historyButton.click();
    return this.page.getByTestId('chat-history-panel');
  }

  async loadSession(sessionIndex: number) {
    const historyPanel = await this.openHistory();
    const sessions = historyPanel.locator('[data-testid^="chat-session-"]');
    await sessions.nth(sessionIndex).click();
  }

  async getMetrics() {
    const metricsButton = this.page.getByTestId('ai-metrics-button');
    await metricsButton.click();
    
    const metricsPanel = this.page.getByTestId('ai-metrics-panel');
    await metricsPanel.waitFor({ state: 'visible' });
    
    return {
      totalTokens: await metricsPanel.getByTestId('total-tokens').textContent(),
      latency: await metricsPanel.getByTestId('latency').textContent(),
      cost: await metricsPanel.getByTestId('cost').textContent(),
      sessionMessages: await metricsPanel.getByTestId('session-messages').textContent(),
    };
  }

  async configureSettings(settings: {
    model?: string;
    temperature?: number;
    maxTokens?: number;
    enableCodeExecution?: boolean;
    enableWebSearch?: boolean;
  }) {
    await this.settingsButton.click();
    const dialog = this.page.getByRole('dialog');
    
    if (settings.model) {
      await dialog.getByTestId('ai-model-select').selectOption(settings.model);
    }
    
    if (settings.temperature !== undefined) {
      await dialog.getByTestId('temperature-slider').fill(settings.temperature.toString());
    }
    
    if (settings.maxTokens !== undefined) {
      await dialog.getByTestId('max-tokens-input').fill(settings.maxTokens.toString());
    }
    
    if (settings.enableCodeExecution !== undefined) {
      const checkbox = dialog.getByTestId('enable-code-execution-checkbox');
      if (settings.enableCodeExecution) {
        await checkbox.check();
      } else {
        await checkbox.uncheck();
      }
    }
    
    if (settings.enableWebSearch !== undefined) {
      const checkbox = dialog.getByTestId('enable-web-search-checkbox');
      if (settings.enableWebSearch) {
        await checkbox.check();
      } else {
        await checkbox.uncheck();
      }
    }
    
    await dialog.getByTestId('save-settings-button').click();
  }

  async getDiagnosticInfo(): Promise<{
    issue: string;
    possibleCauses: string[];
    suggestedActions: string[];
  } | null> {
    const diagnosticCard = this.page.getByTestId('ai-diagnostic-card');
    
    if (!await diagnosticCard.isVisible()) {
      return null;
    }
    
    return {
      issue: await diagnosticCard.getByTestId('diagnostic-issue').textContent() || '',
      possibleCauses: await diagnosticCard.locator('[data-testid^="cause-"]').allTextContents(),
      suggestedActions: await diagnosticCard.locator('[data-testid^="action-"]').allTextContents(),
    };
  }

  async getCodeBlock(): Promise<{
    language: string;
    content: string;
  } | null> {
    const codeBlock = this.page.getByTestId('ai-code-block');
    
    if (!await codeBlock.isVisible()) {
      return null;
    }
    
    const languageClass = await codeBlock.locator('[class*="language-"]').getAttribute('class') || '';
    const language = languageClass.match(/language-(\w+)/)?.[1] || 'text';
    const content = await codeBlock.textContent() || '';
    
    return { language, content };
  }

  async copyCode() {
    const copyButton = this.page.getByTestId('copy-code-button');
    await copyButton.click();
  }

  async getSearchResults(): Promise<Array<{
    title: string;
    url: string;
    relevance: string;
  }>> {
    const searchResults = this.page.getByTestId('ai-search-results');
    
    if (!await searchResults.isVisible()) {
      return [];
    }
    
    const results = searchResults.locator('[data-testid^="search-result-"]');
    const count = await results.count();
    
    const searchResultsData = [];
    for (let i = 0; i < count; i++) {
      const result = results.nth(i);
      searchResultsData.push({
        title: await result.getByTestId('result-title').textContent() || '',
        url: await result.getByTestId('result-url').textContent() || '',
        relevance: await result.getByTestId('result-relevance').textContent() || '',
      });
    }
    
    return searchResultsData;
  }
}