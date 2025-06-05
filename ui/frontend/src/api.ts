import axios from 'axios';
import { Template, TemplateRequest, ExecutionResponse, HealthResponse } from './types';

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
});

export const apiClient = {
  // Get all templates
  async getTemplates(): Promise<Template[]> {
    const response = await api.get<Template[]>('/templates');
    return response.data;
  },

  // Get specific template by path
  async getTemplate(path: string): Promise<Template> {
    const response = await api.get<Template>(`/template/${encodeURIComponent(path)}`);
    return response.data;
  },

  // Execute template with variables
  async executeTemplate(request: TemplateRequest): Promise<ExecutionResponse> {
    const response = await api.post<ExecutionResponse>('/execute', request);
    return response.data;
  },

  // Get health status
  async getHealth(): Promise<HealthResponse> {
    const response = await api.get<HealthResponse>('/health');
    return response.data;
  },
};

export default apiClient; 