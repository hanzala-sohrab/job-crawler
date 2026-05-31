'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/lib/auth';
import { api } from '@/lib/api';
import { Job, SOURCE_CONFIG } from '@/lib/types';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import styles from './page.module.css';

export default function JobDetailPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const params = useParams();
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }
    if (user && params.id) {
      api.getJob(params.id as string)
        .then(setJob)
        .catch(() => router.push('/dashboard'))
        .finally(() => setLoading(false));
    }
  }, [user, authLoading, params.id, router]);

  if (loading || !job) {
    return (
      <div className={styles.page}>
        <div className="container">
          <div className="skeleton" style={{ height: 32, width: '50%', marginBottom: 16 }} />
          <div className="skeleton" style={{ height: 20, width: '30%', marginBottom: 32 }} />
          <div className="skeleton" style={{ height: 200 }} />
        </div>
      </div>
    );
  }

  const sourceConfig = SOURCE_CONFIG[job.source] || { label: job.source, color: '#888', bg: 'rgba(136,136,136,0.15)' };
  const skills: string[] = typeof job.skills === 'string' ? JSON.parse(job.skills) : (job.skills || []);

  return (
    <div className={styles.page}>
      <div className="container">
        <Link href="/dashboard" className={styles.back}>← Back to Dashboard</Link>

        <div className={styles.card}>
          <div className={styles.header}>
            <div>
              <div className={styles.titleRow}>
                <h1 className={styles.title}>{job.title}</h1>
                <span
                  className={`badge badge-source`}
                  style={{
                    color: sourceConfig.color,
                    background: sourceConfig.bg,
                    border: `1px solid ${sourceConfig.color}`,
                  }}
                >
                  {sourceConfig.label}
                </span>
              </div>
              <p className={styles.company}>
                {job.company}
                {job.location && ` · ${job.location}`}
              </p>
            </div>
            {job.url && (
              <a href={job.url} target="_blank" rel="noopener noreferrer" className="btn btn-primary">
                View Original →
              </a>
            )}
          </div>

          {/* Meta */}
          <div className={styles.meta}>
            {job.job_type && <span className={styles.metaTag}>📋 {job.job_type}</span>}
            {job.experience_level && <span className={styles.metaTag}>🎯 {job.experience_level}</span>}
            {job.salary_range && <span className={styles.metaTag}>💰 {job.salary_range}</span>}
            {job.posted_at && (
              <span className={styles.metaTag}>
                📅 {new Date(job.posted_at).toLocaleDateString()}
              </span>
            )}
          </div>

          {/* Skills */}
          {skills.length > 0 && (
            <div className={styles.section}>
              <h2>Required Skills</h2>
              <div className={styles.skillTags}>
                {skills.map((skill: string, i: number) => (
                  <span key={i} className={styles.skillTag}>{skill}</span>
                ))}
              </div>
            </div>
          )}

          {/* Description */}
          {job.description && (
            <div className={styles.section}>
              <h2>Description</h2>
              <div className={styles.description}>
                {job.description.split('\n').map((line, i) => (
                  <p key={i}>{line}</p>
                ))}
              </div>
            </div>
          )}

          {/* Requirements */}
          {job.requirements && (
            <div className={styles.section}>
              <h2>Requirements</h2>
              <div className={styles.description}>
                {job.requirements.split('\n').map((line, i) => (
                  <p key={i}>{line}</p>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
