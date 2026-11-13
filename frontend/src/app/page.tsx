import ConnectionCheck from "@/components/connection-check";
import Link from "next/link";

export default function Home() {
  return (
    <div className="workspace">
      <header className="header">
        <Link className="brand" href="/" aria-label="CLIMS home">
          <span className="brand-mark" aria-hidden="true">
            C
          </span>
          CLIMS<span className="brand-caption">Supply operations</span>
        </Link>
        <a className="reference" href="/docs/">
          API reference <span aria-hidden="true">↗</span>
        </a>
      </header>
      <main>
        <section className="intro">
          <span className="eyebrow">Condom logistics & inventory management</span>
          <h1>
            A connected foundation
            <br />
            for better supply.
          </h1>
          <p>
            Your web and mobile workspace starts here. Connect to the platform and explore the API
            that powers CLIMS.
          </p>
        </section>
        <section className="cards" aria-label="Workspace foundations">
          <article className="card">
            <span className="number">01 / WEB</span>
            <h2>One workspace.</h2>
            <p>A Next.js foundation for bringing supply operations into focus.</p>
            <span className="tag">Web application</span>
          </article>
          <article className="card">
            <span className="number">02 / MOBILE</span>
            <h2>Closer to the field.</h2>
            <p>A React Native foundation for the teams keeping supplies moving.</p>
            <span className="tag">iOS & Android project</span>
          </article>
          <article className="card card-accent">
            <span className="number">03 / PLATFORM</span>
            <h2>Built to connect.</h2>
            <p>Explore authentication, users, and role-based access through the API.</p>
            <a href="/docs/">
              Explore API documentation <span aria-hidden="true">↗</span>
            </a>
          </article>
        </section>
        <ConnectionCheck />
      </main>
      <footer>
        <span>CLIMS</span>
        <span>A foundation for the work ahead.</span>
      </footer>
    </div>
  );
}
