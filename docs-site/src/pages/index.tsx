import React, {type ReactNode} from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './index.module.css';

const paths = [
  ['01', 'Apply your first profile', 'Install the provider, configure credentials, review a plan, and manage your first profile.', '/getting-started/'],
  ['02', 'Understand the provider', 'Follow resource ownership, profile revisions, sensitive state, and service boundaries.', '/development/architecture/'],
  ['03', 'Configure your targets', 'Use native HCL for model providers, custom endpoints, authentication, and existing channels.', '/guides/red-team-testing/'],
  ['04', 'Prove the whole path', 'Check schema contracts, useful plan changes, live lifecycle evidence, and service limits.', '/development/sdk-upgrade-verification/'],
];

export default function Home(): ReactNode {
  return (
    <Layout title="Infrastructure for Prisma AIRS" description="Prisma AIRS Terraform Provider: manage named security profiles, API keys, Model Security groups, and Red Team targets with native HCL.">
      <main>
        <section className={styles.hero} aria-labelledby="hero-title">
          <div className={styles.heroCopy}>
            <p className={styles.eyebrow}>PRISMA AIRS / TERRAFORM PROVIDER</p>
            <h1 id="hero-title">Policy as code.<br /><span>Security under control.</span></h1>
            <p className={styles.lead}>A Terraform provider built for Prisma AIRS. Bring your tenant, configure your credentials, and put security policy and target management to work.</p>
            <div className={styles.actions}>
              <Link className="button button--primary button--lg" to="/getting-started/">Get started →</Link>
              <Link className={styles.secondary} to="/development/architecture/">Explore the architecture ↗</Link>
            </div>
            <p className={styles.platforms}>NATIVE HCL · REVIEWABLE PLANS · MIT LICENSED</p>
          </div>
          <div className={styles.artwork}>
            <img src={useBaseUrl('/img/terraform-logo.png')} alt="Prisma AIRS Terraform shield and prism spectrum" width="1254" height="1254" fetchPriority="high" />
            <div className={styles.pillRow}><span className={styles.pill}>OAuth management</span><span className={styles.pill}>Versioned profiles</span><span className={styles.pill}>Native HCL</span></div>
          </div>
        </section>
        <section className={styles.paths} aria-labelledby="paths-title">
          <div className={styles.sectionIntro}><p className={styles.eyebrow}>FROM FIRST PLAN TO OPERATIONS</p><h2 id="paths-title">A clear path through the platform.</h2><p>Start with the task in front of you. Each guide includes the context, configuration, and checks you need.</p></div>
          <div className={styles.grid}>{paths.map(([number, title, description, to]) => <Link className={styles.path} to={to} key={number}><span className={styles.number}>{number}</span><h3>{title}</h3><p>{description}</p><span className={styles.arrow} aria-hidden="true">↗</span></Link>)}</div>
        </section>
        <section className={styles.quick}><div><p className={styles.eyebrow}>KEEP IT CLOSE</p><h2>Less searching. More doing.</h2><p>Copy the examples for daily work, or look up the exact attributes shipped with the provider.</p></div><div className={styles.actions}><Link className="button button--primary" to="/examples/">Open the examples</Link><Link to="/reference/">Provider reference →</Link></div></section>
      </main>
    </Layout>
  );
}
