import {
  AuthResponse,
  Resume,
  Job,
  JobMatch,
  JobFilter,
  PaginatedResponse,
  SourceInfo,
} from './types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private token: string | null = null;

  constructor() {
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
    }
  }

  getToken(): string | null {
    return this.token;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Don't set Content-Type for FormData (browser sets it with boundary)
    if (!(options.body instanceof FormData)) {
      headers['Content-Type'] = 'application/json';
    }

    const res = await fetch(`${API_URL}${path}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    return res.json();
  }

  // Auth
  async register(email: string, password: string, fullName: string): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, full_name: fullName }),
    });
    this.setToken(res.token);
    return res;
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const res = await this.request<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    this.setToken(res.token);
    return res;
  }

  async getMe() {
    return this.request<AuthResponse['user']>('/api/v1/auth/me');
  }

  // Resume
  async uploadResume(file: File): Promise<Resume> {
    const formData = new FormData();
    formData.append('resume', file);
    return this.request<Resume>('/api/v1/resume/upload', {
      method: 'POST',
      body: formData,
    });
  }

  async getResume(): Promise<Resume> {
    return this.request<Resume>('/api/v1/resume');
  }

  async updateResume(parsedData: unknown): Promise<Resume> {
    return this.request<Resume>('/api/v1/resume', {
      method: 'PUT',
      body: JSON.stringify(parsedData),
    });
  }

  // Jobs
  async getJobs(filter?: JobFilter): Promise<PaginatedResponse<JobMatch>> {
    const params = new URLSearchParams();
    if (filter) {
      Object.entries(filter).forEach(([key, value]) => {
        if (value !== undefined && value !== '') {
          params.set(key, String(value));
        }
      });
    }
    const query = params.toString();
    return this.request<PaginatedResponse<JobMatch>>(
      `/api/v1/jobs${query ? `?${query}` : ''}`
    );
  }

  async getJob(id: string): Promise<Job> {
    return this.request<Job>(`/api/v1/jobs/${id}`);
  }

  async updateJobStatus(jobId: string, status: string): Promise<void> {
    await this.request(`/api/v1/jobs/${jobId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    });
  }

  async getSources(): Promise<SourceInfo[]> {
    return this.request<SourceInfo[]>('/api/v1/jobs/sources');
  }

  async refreshMatches(): Promise<{ message: string }> {
    return this.request<{ message: string }>('/api/v1/jobs/refresh', {
      method: 'POST',
    });
  }
}

// Singleton
export const api = new ApiClient();
