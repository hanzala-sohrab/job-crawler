'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth';
import { api } from '@/lib/api';
import { JobMatch, JobFilter, SOURCE_CONFIG } from '@/lib/types';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import styles from './page.module.css';

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [matches, setMatches] = useState<JobMatch[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [filter, setFilter] = useState<JobFilter>({
    page: 1,
    page_size: 20,
    sort_by: 'relevance',
  });

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
    }
  }, [user, authLoading, router]);

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getJobs(filter);
      setMatches(res.data || []);
      setTotal(res.total);
      setTotalPages(res.total_pages);
    } catch {
      // User might not have a resume yet
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    if (user) fetchJobs();
  }, [user, fetchJobs]);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await api.refreshMatches();
      setTimeout(fetchJobs, 3000);
    } catch {
      // ignore
    } finally {
      setTimeout(() => setRefreshing(false), 3000);
    }
  };

  const handleStatusUpdate = async (jobId: string, status: string) => {
    try {
      await api.updateJobStatus(jobId, status);
      setMatches((prev) =>
        prev.map((m) => (m.job?.id === jobId ? { ...m, status: status as JobMatch['status'] } : m))
      );
    } catch {
      // ignore
    }
  };

  const updateFilter = (updates: Partial<JobFilter>) => {
    setFilter((prev) => ({ ...prev, ...updates, page: 1 }));
  };

  const scoreColor = (score: number) => {
    if (score >= 0.7) return 'var(--success)';
    if (score >= 0.4) return 'var(--warning)';
    return 'var(--text-tertiary)';
  };

  const scoreLabel = (score: number) => {
    if (score >= 0.7) return 'Excellent Match';
    if (score >= 0.4) return 'Good Match';
    return 'Partial Match';
  };

  if (authLoading) return null;

  return (
    <div className={styles.page}>
      <div className="container">
        {/* Header */}
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>Job Dashboard</h1>
            <p className={styles.subtitle}>
              {total > 0 ? `${total} matched jobs found` : 'No matches yet'}
            </p>
          </div>
          <div className={styles.headerActions}>
            <Link href="/resume" className="btn btn-secondary">
              📄 Resume
            </Link>
            <button
              className="btn btn-primary"
              onClick={handleRefresh}
              disabled={refreshing}
            >
              {refreshing ? '⏳ Matching...' : '🔄 Refresh Matches'}
            </button>
          </div>
        </div>

        {/* Filters */}
        <div className={styles.filters}>
          <input
            type="text"
            className={`form-input ${styles.searchInput}`}
            placeholder="Search jobs, companies..."
            value={filter.q || ''}
            onChange={(e) => updateFilter({ q: e.target.value })}
          />
          <select
            className={`form-input ${styles.filterSelect}`}
            value={filter.source || ''}
            onChange={(e) => updateFilter({ source: e.target.value })}
          >
            <option value="">All Sources</option>
            <option value="naukri">Naukri</option>
            <option value="linkedin">LinkedIn</option>
            <option value="hirist">Hirist</option>
            <option value="instahyre">Instahyre</option>
          </select>
          <select
            className={`form-input ${styles.filterSelect}`}
            value={filter.status || ''}
            onChange={(e) => updateFilter({ status: e.target.value })}
          >
            <option value="">All Status</option>
            <option value="new">New</option>
            <option value="saved">Saved</option>
            <option value="applied">Applied</option>
          </select>
          <select
            className={`form-input ${styles.filterSelect}`}
            value={filter.sort_by || 'relevance'}
            onChange={(e) => updateFilter({ sort_by: e.target.value as JobFilter['sort_by'] })}
          >
            <option value="relevance">Sort: Relevance</option>
            <option value="posted_at">Sort: Newest</option>
            <option value="created_at">Sort: Recently Added</option>
          </select>
        </div>

        {/* Job List */}
        {loading ? (
          <div className={styles.jobList}>
            {[...Array(5)].map((_, i) => (
              <div key={i} className={styles.jobCard}>
                <div className="skeleton" style={{ height: 20, width: '60%', marginBottom: 12 }} />
                <div className="skeleton" style={{ height: 16, width: '40%', marginBottom: 8 }} />
                <div className="skeleton" style={{ height: 14, width: '80%' }} />
              </div>
            ))}
          </div>
        ) : matches.length === 0 ? (
          <div className={styles.empty}>
            <span className={styles.emptyIcon}>🎯</span>
            <h3>No job matches yet</h3>
            <p>Upload your resume first, then click &quot;Refresh Matches&quot; to find jobs.</p>
            <Link href="/resume" className="btn btn-primary" style={{ marginTop: 16 }}>
              Upload Resume
            </Link>
          </div>
        ) : (
          <>
            <div className={`${styles.jobList} stagger`}>
              {matches.map((match) => (
                <div key={match.id} className={styles.jobCard}>
                  <div className={styles.jobHeader}>
                    <div className={styles.jobInfo}>
                      <div className={styles.jobTitleRow}>
                        <Link href={`/jobs/${match.job?.id}`} className={styles.jobTitle}>
                          {match.job?.title}
                        </Link>
                        {match.job?.source && (
                          <span
                            className={`badge badge-source ${styles.sourceBadge}`}
                            style={{
                              color: SOURCE_CONFIG[match.job.source]?.color || '#888',
                              background: SOURCE_CONFIG[match.job.source]?.bg || 'rgba(136,136,136,0.15)',
                              borderColor: SOURCE_CONFIG[match.job.source]?.color || '#888',
                            }}
                          >
                            {SOURCE_CONFIG[match.job.source]?.label || match.job.source}
                          </span>
                        )}
                      </div>
                      <p className={styles.jobCompany}>
                        {match.job?.company}
                        {match.job?.location && ` · ${match.job.location}`}
                      </p>
                    </div>
                    <div className={styles.scoreSection}>
                      <div
                        className={styles.scoreCircle}
                        style={{ borderColor: scoreColor(match.relevance_score) }}
                      >
                        <span style={{ color: scoreColor(match.relevance_score) }}>
                          {Math.round(match.relevance_score * 100)}%
                        </span>
                      </div>
                      <span className={styles.scoreLabel} style={{ color: scoreColor(match.relevance_score) }}>
                        {scoreLabel(match.relevance_score)}
                      </span>
                    </div>
                  </div>

                  {/* Matched Skills */}
                  {match.matched_skills?.length > 0 && (
                    <div className={styles.matchedSkills}>
                      {match.matched_skills.slice(0, 6).map((skill, i) => (
                        <span key={i} className={styles.matchedSkillTag}>✓ {skill}</span>
                      ))}
                      {match.matched_skills.length > 6 && (
                        <span className={styles.moreSkills}>+{match.matched_skills.length - 6} more</span>
                      )}
                    </div>
                  )}

                  {/* Meta */}
                  <div className={styles.jobMeta}>
                    <div className={styles.jobTags}>
                      {match.job?.job_type && <span className={styles.metaTag}>{match.job.job_type}</span>}
                      {match.job?.experience_level && <span className={styles.metaTag}>{match.job.experience_level}</span>}
                      {match.job?.salary_range && <span className={styles.metaTag}>💰 {match.job.salary_range}</span>}
                    </div>
                    <div className={styles.jobActions}>
                      {match.status !== 'saved' && (
                        <button
                          className="btn btn-ghost"
                          onClick={() => handleStatusUpdate(match.job!.id, 'saved')}
                        >
                          ⭐ Save
                        </button>
                      )}
                      {match.status !== 'applied' && (
                        <button
                          className="btn btn-ghost"
                          onClick={() => handleStatusUpdate(match.job!.id, 'applied')}
                        >
                          ✅ Applied
                        </button>
                      )}
                      {match.job?.url && (
                        <a
                          href={match.job.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="btn btn-secondary"
                        >
                          View Original →
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className={styles.pagination}>
                <button
                  className="btn btn-ghost"
                  disabled={filter.page === 1}
                  onClick={() => setFilter((p) => ({ ...p, page: (p.page || 1) - 1 }))}
                >
                  ← Previous
                </button>
                <span className={styles.pageInfo}>
                  Page {filter.page} of {totalPages}
                </span>
                <button
                  className="btn btn-ghost"
                  disabled={filter.page === totalPages}
                  onClick={() => setFilter((p) => ({ ...p, page: (p.page || 1) + 1 }))}
                >
                  Next →
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
