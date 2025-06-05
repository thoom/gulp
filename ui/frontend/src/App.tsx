import React, { useState, useEffect } from 'react';
import TemplateList from './components/TemplateList';
import DynamicForm from './components/DynamicForm';
import { Template, ExecutionResponse, HealthResponse } from './types';
import { apiClient } from './api';
import { RefreshCw, Server, Activity, Terminal } from 'lucide-react';
import './App.css';

const App: React.FC = () => {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [loading, setLoading] = useState(true);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [executing, setExecuting] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  // Load initial data
  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [templatesData, healthData] = await Promise.all([
        apiClient.getTemplates(),
        apiClient.getHealth(),
      ]);
      
      setTemplates(templatesData);
      setHealth(healthData);
      setLastRefresh(new Date());
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectTemplate = (template: Template) => {
    setSelectedTemplate(template);
  };

  const handleExecuteTemplate = async (
    variables: Record<string, string>,
    url?: string,
    method?: string
  ): Promise<ExecutionResponse> => {
    if (!selectedTemplate) {
      throw new Error('No template selected');
    }

    setExecuting(true);
    try {
      const response = await apiClient.executeTemplate({
        template_path: selectedTemplate.path,
        variables,
        url,
        method,
      });
      return response;
    } finally {
      setExecuting(false);
    }
  };

  const handleRefresh = () => {
    loadData();
  };

  const formatTimestamp = (date: Date): string => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <div className="header-left">
            <div className="logo">
              <Terminal size={28} className="logo-icon" />
              <h1>Visual GULP</h1>
            </div>
            <div className="subtitle">
              HTTP Client Template Interface
            </div>
          </div>

          <div className="header-right">
            <div className="status-info">
              {health && (
                <>
                  <div className="status-item">
                    <Server size={16} />
                    <span>{health.status}</span>
                  </div>
                  <div className="status-item">
                    <Activity size={16} />
                    <span>{health.templates} templates</span>
                  </div>
                </>
              )}
              <div className="status-item">
                <span className="last-refresh">
                  Updated: {formatTimestamp(lastRefresh)}
                </span>
              </div>
            </div>

            <button
              className="refresh-button"
              onClick={handleRefresh}
              disabled={loading}
              title="Refresh templates"
            >
              <RefreshCw size={16} className={loading ? 'spinning' : ''} />
              Refresh
            </button>
          </div>
        </div>
      </header>

      <main className="app-main">
        <div className="main-content">
          <aside className="sidebar">
            <TemplateList
              templates={templates}
              selectedTemplate={selectedTemplate}
              onSelectTemplate={handleSelectTemplate}
              loading={loading}
            />
          </aside>

          <section className="content">
            <DynamicForm
              template={selectedTemplate}
              onExecute={handleExecuteTemplate}
              loading={executing}
            />
          </section>
        </div>
      </main>

      <footer className="app-footer">
        <div className="footer-content">
          <p>
            GULP v1.0 | 
            Working Directory: <code>{health?.working_dir || 'Loading...'}</code>
          </p>
          <p className="footer-links">
            <a href="https://github.com/thoom/gulp" target="_blank" rel="noopener noreferrer">
              GitHub
            </a>
            {' | '}
            <a href="/api/health" target="_blank" rel="noopener noreferrer">
              API Health
            </a>
          </p>
        </div>
      </footer>
    </div>
  );
};

export default App; 