/* eslint-disable */
// TrialDetailDrawer — slides in from the right when you click a trial card.
// Shows trial info, walkthrough timeline, entry CTA.

function TrialDetailDrawer({ trial, onClose }) {
  if (!trial) return null;
  return (
    <div style={{ position: 'fixed', inset: 0, zIndex: 50 }} role="dialog" aria-modal="true">
      <div onClick={onClose}
           style={{ position: 'absolute', inset: 0, background: 'rgb(0 0 0 / 0.40)', animation: 'fade 200ms both' }} />
      <aside
        className="fade-up"
        style={{
          position: 'absolute', top: 0, right: 0, bottom: 0,
          width: 'min(560px, 100%)',
          background: 'var(--mist-50)',
          boxShadow: 'var(--shadow-lg)',
          overflow: 'auto',
          padding: '40px 32px',
          display: 'flex', flexDirection: 'column', gap: 24,
        }}
        data-event={trial.event}
      >
        <div className="flex justify-between items-start gap-4">
          <div className="flex-col gap-2">
            <Eyebrow>{trial.event[0].toUpperCase() + trial.event.slice(1)} · Trial</Eyebrow>
            <h2 className="h2" style={{ fontSize: '2rem' }}>{trial.title}</h2>
          </div>
          <button onClick={onClose} aria-label="Close" style={{ background: 'transparent', border: 0, padding: 8, cursor: 'pointer', color: 'var(--mist-700)' }}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path d="M3 3 L13 13 M13 3 L3 13" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
            </svg>
          </button>
        </div>

        <PhotoPlaceholder kind={trial.event === 'tracking' ? 'track' : trial.event === 'detection' ? 'det' : trial.event === 'obedience' ? 'ob' : 'prot'} height="220px" maxHeight={220} caption={`${trial.loc}`} eventTint={trial.event} play={false} />

        <div className="flex-col gap-4">
          <div className="flex items-center gap-3 body">
            <Icon name="calendar" /> <span style={{ color: 'var(--mist-950)', fontWeight: 500 }}>{trial.date}</span>
          </div>
          <div className="flex items-center gap-3 body">
            <Icon name="clock" /> {trial.time}
          </div>
          <div className="flex items-center gap-3 body">
            <Icon name="map-pin" /> {trial.loc}
          </div>
          <div className="flex items-center gap-3 body">
            <Icon name="user-circle" /> Judge: Anya Brunner · DV
          </div>
        </div>

        <div className="flex-col gap-3" style={{ padding: 16, background: 'rgb(from var(--mist-950) r g b / 0.025)', borderRadius: 12 }}>
          <div className="label-xs">Day schedule</div>
          <div className="flex-col gap-2 body-sm" style={{ color: 'var(--mist-950)' }}>
            <div className="flex justify-between"><span>05.30</span><span>Gate opens · check-in</span></div>
            <div className="flex justify-between"><span>06.30</span><span>Walkthrough</span></div>
            <div className="flex justify-between"><span>07.30</span><span>First run</span></div>
            <div className="flex justify-between"><span>15.00</span><span>Awards · brief debrief</span></div>
          </div>
        </div>

        <div className="flex justify-between items-center" style={{ paddingTop: 16, borderTop: '1px solid var(--mist-200)' }}>
          <div className="body-sm"><b style={{ color: 'var(--mist-950)' }}>{trial.spots}</b> of {trial.cap} spots filled</div>
          <Button variant="event" size="lg">
            {trial.status === 'closed' ? 'Brief PDF' : trial.status === 'wait' ? 'Join waitlist' : 'Enter trial'} <Icon name="arrow-right" />
          </Button>
        </div>

        <p className="body-sm" style={{ color: 'var(--mist-600)' }}>
          Entries close 72 hours before walkthrough. Briefs go out at 18.00 the night before.
        </p>
      </aside>
    </div>
  );
}

window.TrialDetailDrawer = TrialDetailDrawer;
