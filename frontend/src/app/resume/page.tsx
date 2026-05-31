'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import { useAuth } from '@/lib/auth';
import { api } from '@/lib/api';
import { Resume, ParsedResume } from '@/lib/types';
import { useRouter } from 'next/navigation';
import styles from './page.module.css';

export default function ResumePage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [resume, setResume] = useState<Resume | null>(null);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }
    if (user) {
      api.getResume()
        .then(setResume)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [user, authLoading, router]);

  const handleUpload = useCallback(async (file: File) => {
    setError('');
    const ext = file.name.split('.').pop()?.toLowerCase();
    if (!['pdf', 'docx', 'doc'].includes(ext || '')) {
      setError('Only PDF and DOCX files are supported');
      return;
    }
    if (file.size > 10 * 1024 * 1024) {
      setError('File must be smaller than 10MB');
      return;
    }

    setUploading(true);
    try {
      const result = await api.uploadResume(file);
      setResume(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files?.[0];
    if (file) handleUpload(file);
  }, [handleUpload]);

  const parsed: ParsedResume | null = resume?.parsed_data || null;

  if (authLoading || loading) {
    return (
      <div className={styles.page}>
        <div className="container">
          <div className={styles.skeleton} style={{ height: 200 }} />
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className="container">
        <h1 className={styles.title}>Your Resume</h1>

        {/* Upload Zone */}
        <div
          className={`${styles.dropzone} ${dragOver ? styles.dropzoneActive : ''} ${uploading ? styles.dropzoneUploading : ''}`}
          onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf,.docx,.doc"
            onChange={(e) => e.target.files?.[0] && handleUpload(e.target.files[0])}
            style={{ display: 'none' }}
          />
          {uploading ? (
            <>
              <div className={styles.spinner} />
              <p className={styles.dropText}>Parsing your resume with AI...</p>
              <p className={styles.dropHint}>This may take a few seconds</p>
            </>
          ) : (
            <>
              <span className={styles.dropIcon}>📄</span>
              <p className={styles.dropText}>
                {resume ? 'Upload a new resume' : 'Drop your resume here or click to browse'}
              </p>
              <p className={styles.dropHint}>PDF or DOCX, max 10MB</p>
            </>
          )}
        </div>

        {error && <div className={styles.error}>{error}</div>}

        {/* Parsed Resume View */}
        {parsed && (
          <div className={styles.parsedSection}>
            <div className={styles.parsedHeader}>
              <h2>Parsed Resume</h2>
              {resume?.file_name && (
                <span className={styles.fileName}>📎 {resume.file_name}</span>
              )}
            </div>

            <div className={styles.parsedGrid}>
              {/* Personal Info */}
              <div className={styles.infoCard}>
                <h3>Personal Info</h3>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>Name</span>
                  <span>{parsed.full_name || '—'}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>Email</span>
                  <span>{parsed.email || '—'}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>Phone</span>
                  <span>{parsed.phone || '—'}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>Location</span>
                  <span>{parsed.location || '—'}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>Experience</span>
                  <span>{parsed.total_experience_years ? `${parsed.total_experience_years} years` : '—'}</span>
                </div>
              </div>

              {/* Skills */}
              <div className={styles.infoCard}>
                <h3>Skills ({parsed.skills?.length || 0})</h3>
                <div className={styles.skillTags}>
                  {parsed.skills?.map((skill, i) => (
                    <span key={i} className={styles.skillTag}>{skill}</span>
                  ))}
                </div>
              </div>

              {/* Experience */}
              {parsed.experience?.length > 0 && (
                <div className={`${styles.infoCard} ${styles.fullWidth}`}>
                  <h3>Experience</h3>
                  {parsed.experience.map((exp, i) => (
                    <div key={i} className={styles.expItem}>
                      <div className={styles.expHeader}>
                        <strong>{exp.title}</strong>
                        <span className={styles.expCompany}>@ {exp.company}</span>
                      </div>
                      <span className={styles.expDate}>
                        {exp.start_date} — {exp.end_date || 'Present'}
                      </span>
                      {exp.description && <p className={styles.expDesc}>{exp.description}</p>}
                    </div>
                  ))}
                </div>
              )}

              {/* Education */}
              {parsed.education?.length > 0 && (
                <div className={styles.infoCard}>
                  <h3>Education</h3>
                  {parsed.education.map((edu, i) => (
                    <div key={i} className={styles.expItem}>
                      <strong>{edu.degree} in {edu.field}</strong>
                      <p className={styles.expDate}>{edu.institution} • {edu.year}</p>
                    </div>
                  ))}
                </div>
              )}

              {/* Preferred Titles */}
              {parsed.preferred_job_titles?.length > 0 && (
                <div className={styles.infoCard}>
                  <h3>Preferred Job Titles</h3>
                  <div className={styles.skillTags}>
                    {parsed.preferred_job_titles.map((t, i) => (
                      <span key={i} className={styles.skillTag}>{t}</span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <div className={styles.actions}>
              <button
                className="btn btn-primary"
                onClick={() => router.push('/dashboard')}
              >
                View Job Matches →
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
