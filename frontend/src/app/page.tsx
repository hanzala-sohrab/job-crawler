'use client';

import Link from 'next/link';
import styles from './page.module.css';

export default function HomePage() {
  return (
    <div className={styles.page}>
      {/* Hero Section */}
      <section className={styles.hero}>
        <div className={`container ${styles.heroContent}`}>
          <div className={styles.badge}>
            Job Matching Platform
          </div>
          <h1 className={styles.title}>
            Your Resume.
            <br />
            Every Job. One Platform.
          </h1>
          <p className={styles.subtitle}>
            Upload your resume and let the system find the best job matches across
            multiple job boards and company career pages — all in one place.
          </p>
          <div className={styles.cta}>
            <Link href="/register" className="btn btn-primary" style={{ padding: '16px 32px', fontSize: '1rem' }}>
              Get Started Free →
            </Link>
            <Link href="/login" className="btn btn-secondary" style={{ padding: '16px 32px', fontSize: '1rem' }}>
              Sign In
            </Link>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className={styles.features}>
        <div className="container">
          <h2 className={styles.sectionTitle}>How It Works</h2>
          <div className={styles.featureGrid}>
            {[
              {
                icon: '📄',
                title: 'Upload Resume',
                desc: 'Upload your PDF or DOCX resume. We extract your skills, experience, and preferences automatically.',
              },
              {
                icon: '⚡',
                title: 'Data Extraction',
                desc: 'We parse your resume into structured data — skills, experience, education, and more.',
              },
              {
                icon: '🔍',
                title: 'Multi-Source Search',
                desc: 'We crawl multiple sources to find jobs matching your profile.',
              },
              {
                icon: '🎯',
                title: 'Smart Matching',
                desc: 'Our scoring engine ranks jobs by skill match, title relevance, experience fit, and location.',
              },
              {
                icon: '📊',
                title: 'Track & Apply',
                desc: 'Save jobs, mark as applied, and track your applications — all from a unified dashboard.',
              },
              {
                icon: '🔔',
                title: 'Stay Updated',
                desc: 'New jobs are scraped regularly. Refresh your matches anytime to see the latest opportunities.',
              },
            ].map((feature, i) => (
              <div key={i} className={styles.featureCard}>
                <span className={styles.featureIcon}>{feature.icon}</span>
                <h3 className={styles.featureTitle}>{feature.title}</h3>
                <p className={styles.featureDesc}>{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Sources */}
      <section className={styles.sources}>
        <div className="container">
          <h2 className={styles.sectionTitle}>Jobs From Everywhere</h2>
          <div className={styles.sourceLogos}>
            {['LinkedIn', 'Naukri', 'Hirist', 'Instahyre', 'Company Pages'].map((source) => (
              <div key={source} className={styles.sourceBadge}>
                {source}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className={styles.footer}>
        <div className="container">
          <p>© 2026 JobCrawler. Built with Go + Next.js.</p>
        </div>
      </footer>
    </div>
  );
}
