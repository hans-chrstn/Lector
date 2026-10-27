import { ArrowDownToLine, HardDrive, ShieldCheck } from 'lucide-react';
export default function Settings() {
  return (
    <div className="page">
      <header className="page-heading">
        <div>
          <p className="eyebrow">YOUR SPACE, YOUR TERMS</p>
          <h1>
            Keep it personal<span className="heading-dot">.</span>
          </h1>
          <p className="page-description">A library you can take care of, and take with you.</p>
        </div>
      </header>
      <div className="settings-stack">
        <section className="settings-card">
          <HardDrive size={25} />
          <div>
            <h2>Your library lives here</h2>
            <p>
              Titles, progress, and notes are saved to your Lector server. Imported files stay in
              its data folder. There is no cloud account or background sync.
            </p>
          </div>
        </section>
        <section className="settings-card">
          <ArrowDownToLine size={25} />
          <div>
            <h2>Take your library with you</h2>
            <p>
              Download your titles, notes, collections, and progress as JSON. Imported files and
              activity history are not included. Back up your data folder for a complete copy.
            </p>
            <a className="button secondary" href="/api/export" download="lector-library.json">
              <ArrowDownToLine size={16} />
              Export library
            </a>
          </div>
        </section>
        <section className="settings-card">
          <ShieldCheck size={25} />
          <div>
            <h2>A home for your own collection</h2>
            <p>
              Lector organizes the files and titles you bring. You control which personal extensions
              are enabled and what they can access.
            </p>
          </div>
        </section>
      </div>
    </div>
  );
}
