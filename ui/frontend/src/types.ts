export interface Template {
  path: string;
  name: string;
  content: string;
  variables: string[];
  folder: string;
  size: number;
  modified: string;
  is_valid: boolean;
  error?: string;
}

export interface TemplateRequest {
  template_path: string;
  variables: Record<string, string>;
  url?: string;
  method?: string;
}

export interface ExecutionResponse {
  success: boolean;
  status_code?: number;
  headers?: Record<string, string>;
  body: string;
  error?: string;
  duration: number;
  request_url: string;
  request_headers?: Record<string, string>;
}

export interface HealthResponse {
  status: string;
  templates: number;
  working_dir: string;
  timestamp: string;
}

export interface TemplateGroup {
  folder: string;
  templates: Template[];
  count: number;
} 