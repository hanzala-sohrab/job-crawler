// API client types matching the Go backend models

export interface User {
  id: string;
  email: string;
  full_name: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ParsedResume {
  full_name: string;
  email: string;
  phone: string;
  location: string;
  summary: string;
  total_experience_years: number;
  skills: string[];
  experience: Experience[];
  education: Education[];
  certifications: string[];
  preferred_job_titles: string[];
  preferred_locations: string[];
}

export interface Experience {
  company: string;
  title: string;
  start_date: string;
  end_date: string;
  description: string;
  technologies: string[];
}

export interface Education {
  institution: string;
  degree: string;
  field: string;
  year: number;
}

export interface Resume {
  id: string;
  user_id: string;
  file_path?: string;
  file_name?: string;
  parsed_data?: ParsedResume;
  created_at: string;
  updated_at: string;
}

export interface Job {
  id: string;
  external_id?: string;
  source: string;
  title: string;
  company: string;
  location: string;
  description: string;
  requirements?: string;
  salary_range?: string;
  job_type?: string;
  experience_level?: string;
  skills: string[];
  url: string;
  posted_at?: string;
  scraped_at: string;
  is_active: boolean;
  created_at: string;
}

export interface JobMatch {
  id: string;
  user_id: string;
  job_id: string;
  relevance_score: number;
  matched_skills: string[];
  status: 'new' | 'saved' | 'applied' | 'hidden';
  created_at: string;
  job?: Job;
}

export interface JobFilter {
  source?: string;
  location?: string;
  job_type?: string;
  experience_level?: string;
  status?: string;
  q?: string;
  skills?: string;
  page?: number;
  page_size?: number;
  sort_by?: 'relevance' | 'posted_at' | 'created_at';
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface SourceInfo {
  source: string;
  count: number;
}

// Source display config
export const SOURCE_CONFIG: Record<string, { label: string; color: string; bg: string }> = {
  naukri: { label: 'Naukri', color: '#00b8a9', bg: 'rgba(0,184,169,0.15)' },
  linkedin: { label: 'LinkedIn', color: '#0a66c2', bg: 'rgba(10,102,194,0.15)' },
  hirist: { label: 'Hirist', color: '#ff6b35', bg: 'rgba(255,107,53,0.15)' },
  instahyre: { label: 'Instahyre', color: '#7c3aed', bg: 'rgba(124,58,237,0.15)' },
  company: { label: 'Company', color: '#059669', bg: 'rgba(5,150,105,0.15)' },
};
