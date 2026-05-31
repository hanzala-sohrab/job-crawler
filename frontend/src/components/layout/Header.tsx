'use client';

import { useAuth } from '@/lib/auth';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import styles from './Header.module.css';

export default function Header() {
  const { user, logout, loading } = useAuth();
  const pathname = usePathname();

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <Link href="/" className={styles.logo}>
          <span className={styles.logoIcon}>⚡</span>
          <span className={styles.logoText}>JobCrawler</span>
        </Link>

        <nav className={styles.nav}>
          {!loading && user ? (
            <>
              <Link
                href="/dashboard"
                className={`${styles.navLink} ${pathname === '/dashboard' ? styles.active : ''}`}
              >
                Dashboard
              </Link>
              <Link
                href="/resume"
                className={`${styles.navLink} ${pathname === '/resume' ? styles.active : ''}`}
              >
                Resume
              </Link>
              <div className={styles.userSection}>
                <span className={styles.userName}>{user.full_name || user.email}</span>
                <button onClick={logout} className={styles.logoutBtn}>
                  Logout
                </button>
              </div>
            </>
          ) : !loading ? (
            <>
              <Link href="/login" className="btn btn-ghost">
                Sign In
              </Link>
              <Link href="/register" className="btn btn-primary">
                Get Started
              </Link>
            </>
          ) : null}
        </nav>
      </div>
    </header>
  );
}
