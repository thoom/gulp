import React, { useState, useEffect, useRef } from 'react';
import { Play, Settings, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { Template, ExecutionResponse } from '../types';

interface DynamicFormProps {
  template: Template | null;
  onExecute: (variables: Record<string, string>, url?: string, method?: string) => Promise<ExecutionResponse>;
  loading: boolean;
}

const DynamicForm: React.FC<DynamicFormProps> = ({
  template,
  onExecute,
  loading,
}) => {
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [customUrl, setCustomUrl] = useState('');
  const [customMethod, setCustomMethod] = useState('GET');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [lastExecution, setLastExecution] = useState<ExecutionResponse | null>(null);
  const [executing, setExecuting] = useState(false);
  const [activeTab, setActiveTab] = useState<'request' | 'response' | 'body'>('body');
  
  // Use ref to track previous variables to avoid dependency issues
  const prevVariablesRef = useRef<Record<string, string>>({});

  // Reset form when template changes
  useEffect(() => {
    if (template) {
      const newVariables: Record<string, string> = {};
      template.variables.forEach(variable => {
        // Preserve existing value if it exists
        newVariables[variable] = prevVariablesRef.current[variable] || '';
      });
      setVariables(newVariables);
      setCustomUrl('');
      setCustomMethod('GET');
      setShowAdvanced(false);
      setActiveTab('body'); // Reset to first tab (Response Body)
    }
  }, [template]);

  // Update ref when variables change
  useEffect(() => {
    prevVariablesRef.current = variables;
  }, [variables]);

  const handleVariableChange = (variable: string, value: string) => {
    setVariables(prev => ({
      ...prev,
      [variable]: value,
    }));
  };

  const handleExecute = async () => {
    if (!template) return;
    
    setExecuting(true);
    try {
      const response = await onExecute(
        variables,
        customUrl || undefined,
        customMethod || undefined
      );
      setLastExecution(response);
      // Auto-switch to appropriate tab based on result
      if (response.success && response.body) {
        setActiveTab('body');
      } else if (response.success && response.request_headers && Object.keys(response.request_headers).length > 0) {
        setActiveTab('request');
      }
    } catch (error) {
      console.error('Execution failed:', error);
      setLastExecution({
        success: false,
        body: '',
        error: error instanceof Error ? error.message : 'Unknown error',
        duration: 0,
        request_url: customUrl || '',
      });
    } finally {
      setExecuting(false);
    }
  };

  const isFormValid = () => {
    if (!template) return false;
    return template.variables.every(variable => variables[variable]?.trim() !== '');
  };

  const getVariableHint = (variable: string): string => {
    const hints: Record<string, string> = {
      environment: 'e.g., dev, staging, prod',
      endpoint: 'e.g., users, orders, products',
      token: 'API authentication token',
      api_token: 'API authentication token',
      url: 'Full URL endpoint',
      method: 'HTTP method (GET, POST, PUT, DELETE)',
      timeout: 'Request timeout in seconds',
      port: 'Port number (e.g., 8080, 3000)',
      host: 'Hostname or IP address',
      version: 'API version (e.g., v1, v2)',
      id: 'Resource identifier',
      user_id: 'User identifier',
      client_id: 'Client identifier',
      domain: 'Domain name',
      subdomain: 'Subdomain prefix',
    };
    
    return hints[variable.toLowerCase()] || `Value for ${variable}`;
  };

  const formatResponseBody = (body: string): string => {
    // Try to parse and pretty-print JSON
    try {
      const parsed = JSON.parse(body);
      return JSON.stringify(parsed, null, 2);
    } catch {
      // If not valid JSON, return as-is
      return body;
    }
  };

  // Check if the response is an HTTP error (4xx or 5xx status codes)
  const isHttpError = (execution: ExecutionResponse): boolean => {
    return execution.success && execution.status_code !== undefined && execution.status_code >= 400;
  };

  // Get the overall result status (considering both execution success and HTTP status)
  const getResultStatus = (execution: ExecutionResponse): 'success' | 'error' => {
    if (!execution.success || isHttpError(execution)) {
      return 'error';
    }
    return 'success';
  };

  if (!template) {
    return (
      <div className="dynamic-form empty">
        <div className="empty-state">
          <Settings size={48} className="text-gray-400" />
          <h3>Select a Template</h3>
          <p>Choose a template from the list to configure and execute it.</p>
        </div>
      </div>
    );
  }

  if (!template.is_valid) {
    return (
      <div className="dynamic-form error">
        <div className="error-state">
          <AlertCircle size={48} className="text-red-500" />
          <h3>Invalid Template</h3>
          <p>This template has syntax errors and cannot be executed:</p>
          <div className="error-details">
            <code>{template.error}</code>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="dynamic-form">
      <div className="form-header">
        <h2>Configure Template</h2>
        <p className="template-name">{template.name}</p>
      </div>

      <form onSubmit={(e) => { e.preventDefault(); handleExecute(); }}>
        {template.variables.length > 0 ? (
          <div className="variables-section">
            <h3>Template Variables</h3>
            <div className="variables-grid">
              {template.variables.map((variable) => (
                <div key={variable} className="variable-field">
                  <label htmlFor={`var-${variable}`}>
                    {variable}
                    <span className="required">*</span>
                  </label>
                  <input
                    id={`var-${variable}`}
                    type="text"
                    value={variables[variable] || ''}
                    onChange={(e) => handleVariableChange(variable, e.target.value)}
                    placeholder={getVariableHint(variable)}
                    required
                  />
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="no-variables">
            <p>This template doesn't require any variables.</p>
          </div>
        )}

        <div className="advanced-section">
          <button
            type="button"
            className="toggle-advanced"
            onClick={() => setShowAdvanced(!showAdvanced)}
          >
            <Settings size={16} />
            Advanced Options
          </button>

          {showAdvanced && (
            <div className="advanced-options">
              <div className="field-group">
                <label htmlFor="custom-url">Custom URL (optional)</label>
                <input
                  id="custom-url"
                  type="url"
                  value={customUrl}
                  onChange={(e) => setCustomUrl(e.target.value)}
                  placeholder="https://api.example.com/endpoint"
                />
              </div>

              <div className="field-group">
                <label htmlFor="custom-method">HTTP Method</label>
                <select
                  id="custom-method"
                  value={customMethod}
                  onChange={(e) => setCustomMethod(e.target.value)}
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                  <option value="PATCH">PATCH</option>
                </select>
              </div>
            </div>
          )}
        </div>

        <div className="form-actions">
          <button
            type="submit"
            className="execute-button"
            disabled={!isFormValid() || executing}
          >
            {executing ? (
              <>
                <div className="button-spinner"></div>
                Executing...
              </>
            ) : (
              <>
                <Play size={16} />
                Execute Template
              </>
            )}
          </button>
        </div>
      </form>

      {lastExecution && (
        <div className={`execution-result ${getResultStatus(lastExecution)}`}>
          <div className="result-header">
            {getResultStatus(lastExecution) === 'success' ? (
              <CheckCircle size={20} className="text-green-500" />
            ) : (
              <AlertCircle size={20} className="text-red-500" />
            )}
            <h3>Execution Result</h3>
            <div className="result-meta">
              <Clock size={14} />
              <span>{lastExecution.duration.toFixed(2)}s</span>
            </div>
          </div>

          {lastExecution.success && lastExecution.status_code && (
            <div className="status-info">
              <span className={`status-code ${lastExecution.status_code >= 400 ? 'error' : 'success'}`}>
                {lastExecution.status_code}
              </span>
              <span className="request-url">{lastExecution.request_url}</span>
            </div>
          )}

          {(lastExecution.error || isHttpError(lastExecution)) && (
            <div className="error-message">
              <strong>Error:</strong> {
                lastExecution.error || 
                (isHttpError(lastExecution) ? `HTTP ${lastExecution.status_code} Error` : 'Unknown error')
              }
            </div>
          )}

          {/* Tabbed Interface for Request/Response Data */}
          <div className="result-tabs">
            <div className="tab-buttons">
              <button
                className={`tab-button ${activeTab === 'body' ? 'active' : ''}`}
                onClick={() => setActiveTab('body')}
                disabled={!lastExecution.body}
              >
                Response Body
              </button>
              <button
                className={`tab-button ${activeTab === 'request' ? 'active' : ''}`}
                onClick={() => setActiveTab('request')}
                disabled={!lastExecution.request_headers || Object.keys(lastExecution.request_headers).length === 0}
              >
                Request Headers
                {lastExecution.request_headers && Object.keys(lastExecution.request_headers).length > 0 && (
                  <span className="tab-badge">{Object.keys(lastExecution.request_headers).length}</span>
                )}
              </button>
              <button
                className={`tab-button ${activeTab === 'response' ? 'active' : ''}`}
                onClick={() => setActiveTab('response')}
                disabled={!lastExecution.headers || Object.keys(lastExecution.headers).length === 0}
              >
                Response Headers
                {lastExecution.headers && Object.keys(lastExecution.headers).length > 0 && (
                  <span className="tab-badge">{Object.keys(lastExecution.headers).length}</span>
                )}
              </button>
            </div>

            <div className="tab-content">
              {activeTab === 'request' && lastExecution.request_headers && Object.keys(lastExecution.request_headers).length > 0 && (
                <div className="tab-pane request-headers-tab">
                  <div className="headers-list">
                    {Object.entries(lastExecution.request_headers).map(([key, value]) => (
                      <div key={key} className="header-item">
                        <strong>{key}:</strong> {value}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {activeTab === 'response' && lastExecution.headers && Object.keys(lastExecution.headers).length > 0 && (
                <div className={`tab-pane response-headers-tab ${isHttpError(lastExecution) ? 'error' : ''}`}>
                  <div className="headers-list">
                    {Object.entries(lastExecution.headers).map(([key, value]) => (
                      <div key={key} className="header-item">
                        <strong>{key}:</strong> {value}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {activeTab === 'body' && lastExecution.body && (
                <div className={`tab-pane response-body-tab ${isHttpError(lastExecution) ? 'error' : ''}`}>
                  <pre><code>{formatResponseBody(lastExecution.body)}</code></pre>
                </div>
              )}

              {/* Empty state for tabs with no content */}
              {((activeTab === 'request' && (!lastExecution.request_headers || Object.keys(lastExecution.request_headers).length === 0)) ||
                (activeTab === 'response' && (!lastExecution.headers || Object.keys(lastExecution.headers).length === 0)) ||
                (activeTab === 'body' && !lastExecution.body)) && (
                <div className="tab-pane empty-tab">
                  <p>No {activeTab === 'request' ? 'request headers' : activeTab === 'response' ? 'response headers' : 'response body'} available</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default DynamicForm; 