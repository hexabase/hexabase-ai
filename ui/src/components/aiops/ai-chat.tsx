'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Send, 
  AlertCircle, 
  Download, 
  Trash2, 
  Zap,
  DollarSign,
  Shield,
  Activity,
  ChevronRight,
  Loader2
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  suggestions?: string[];
}

interface AIChatProps {
  workspaceId: string;
}

interface Suggestion {
  id: string;
  type: string;
  title: string;
  description: string;
  priority: string;
  estimated_savings?: string;
}

export function AIChat({ workspaceId }: AIChatProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();

  useEffect(() => {
    fetchSuggestions();
  }, [workspaceId]);

  useEffect(() => {
    // Scroll to bottom when new messages are added
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  }, [messages]);

  const fetchSuggestions = async () => {
    try {
      const response = await apiClient.aiops.getSuggestions();
      setSuggestions(response.suggestions);
    } catch (err) {
      console.error('Failed to fetch suggestions:', err);
    }
  };

  const sendMessage = async (messageText: string) => {
    if (!messageText.trim() || loading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: messageText,
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setLoading(true);
    setError(null);

    try {
      const response = await apiClient.aiops.chat({
        message: messageText,
        context: { workspace_id: workspaceId }
      });

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: response.message,
        timestamp: new Date(),
        suggestions: response.suggestions
      };

      setMessages(prev => [...prev, assistantMessage]);
    } catch (err) {
      setError('Failed to get AI response');
      toast({
        title: 'Error',
        description: 'Failed to get AI response. Please try again.',
        variant: 'destructive'
      });
    } finally {
      setLoading(false);
    }
  };

  const handleQuickAction = (action: string) => {
    sendMessage(action);
  };

  const analyzeMetrics = async () => {
    setLoading(true);
    try {
      const response = await apiClient.aiops.analyzeMetrics(workspaceId);
      
      const analysisMessage: Message = {
        id: Date.now().toString(),
        role: 'assistant',
        content: response.analysis.summary,
        timestamp: new Date()
      };

      setMessages(prev => [...prev, analysisMessage]);

      // Add findings as separate messages
      if (response.analysis.findings) {
        response.analysis.findings.forEach((finding: any) => {
          const findingMessage = `${finding.severity.toUpperCase()}: ${finding.description}`;
          toast({
            title: finding.type,
            description: findingMessage,
            variant: finding.severity === 'warning' ? 'destructive' : 'default'
          });
        });
      }
    } catch (err) {
      setError('Failed to analyze metrics');
    } finally {
      setLoading(false);
    }
  };

  const exportChat = () => {
    const chatHistory = messages.map(msg => 
      `[${msg.timestamp.toISOString()}] ${msg.role.toUpperCase()}: ${msg.content}`
    ).join('\n\n');

    const blob = new Blob([chatHistory], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `chat-history-${new Date().toISOString()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast({
      title: 'Success',
      description: 'Chat history exported'
    });
  };

  const clearChat = () => {
    if (window.confirm('Are you sure you want to clear the chat history?')) {
      setMessages([]);
    }
  };

  const quickActions = [
    { icon: Zap, label: 'Check resource usage', action: 'Check resource usage' },
    { icon: DollarSign, label: 'Analyze costs', action: 'Analyze costs' },
    { icon: Shield, label: 'Security scan', action: 'Security scan' },
    { icon: Activity, label: 'Performance tips', action: 'Performance tips' }
  ];

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Zap className="h-5 w-5" />
          AI Assistant
        </CardTitle>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowSuggestions(!showSuggestions)}
          >
            View Suggestions
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={analyzeMetrics}
            disabled={loading}
          >
            Analyze Metrics
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={exportChat}
            disabled={messages.length === 0}
          >
            <Download className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={clearChat}
            disabled={messages.length === 0}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col p-0">
        {showSuggestions && suggestions.length > 0 && (
          <div className="border-b p-4 space-y-2">
            <h4 className="font-medium text-sm">AI Suggestions</h4>
            {suggestions.map(suggestion => (
              <div
                key={suggestion.id}
                className="flex items-start gap-3 p-3 bg-muted rounded-lg hover:bg-muted/80 cursor-pointer"
                onClick={() => sendMessage(`Tell me more about: ${suggestion.title}`)}
              >
                <Badge variant={suggestion.priority === 'high' ? 'destructive' : 'secondary'}>
                  {suggestion.type}
                </Badge>
                <div className="flex-1">
                  <p className="font-medium text-sm">{suggestion.title}</p>
                  <p className="text-sm text-muted-foreground">{suggestion.description}</p>
                  {suggestion.estimated_savings && (
                    <p className="text-sm text-green-600 mt-1">{suggestion.estimated_savings}</p>
                  )}
                </div>
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </div>
            ))}
          </div>
        )}

        <ScrollArea className="flex-1 p-4" ref={scrollAreaRef}>
          <div className="space-y-4">
            {messages.length === 0 && (
              <div className="text-center text-muted-foreground py-8">
                <p>Ask me about your infrastructure!</p>
                <p className="text-sm mt-2">I can help with resource optimization, cost analysis, and more.</p>
              </div>
            )}

            {messages.map(message => (
              <div
                key={message.id}
                className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[80%] rounded-lg p-3 ${
                    message.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted'
                  }`}
                >
                  <p className="text-sm">{message.content}</p>
                  {message.suggestions && message.suggestions.length > 0 && (
                    <div className="mt-2 pt-2 border-t border-border/50">
                      <p className="text-xs font-medium mb-1">Suggestions:</p>
                      <ul className="text-xs space-y-1">
                        {message.suggestions.map((suggestion, idx) => (
                          <li key={idx}>• {suggestion}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              </div>
            ))}

            {loading && (
              <div className="flex justify-start">
                <div className="bg-muted rounded-lg p-3">
                  <Loader2 className="h-4 w-4 animate-spin" data-testid="ai-loading" />
                </div>
              </div>
            )}

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </div>
        </ScrollArea>

        <div className="border-t p-4 space-y-3">
          <div className="flex gap-2">
            {quickActions.map(({ icon: Icon, label, action }) => (
              <Button
                key={label}
                variant="outline"
                size="sm"
                onClick={() => handleQuickAction(action)}
                disabled={loading}
              >
                <Icon className="h-4 w-4 mr-1" />
                {label}
              </Button>
            ))}
          </div>

          <div className="flex gap-2">
            <Input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && sendMessage(input)}
              placeholder="Ask AI about your infrastructure..."
              disabled={loading}
            />
            <Button onClick={() => sendMessage(input)} disabled={loading || !input.trim()}>
              <Send className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>

    </Card>
  );
}