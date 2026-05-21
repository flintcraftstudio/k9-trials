/* eslint-disable */
// TrialList — used on the home page (top 3 upcoming) and the /trials page (full list).

const TRIALS = [
  { id: 't1', event: 'detection',  title: 'Spring Open · Eastern Region',     date: 'Sat, May 24',  time: '06.30 walkthrough', loc: 'Reston, VA',     spots: 18, cap: 24, status: 'open',   close: '6 days' },
  { id: 't2', event: 'tracking',   title: 'May Tracking Trial · Skyline',     date: 'Sun, May 25',  time: '05.00 first track',  loc: 'Front Royal, VA',spots: 12, cap: 12, status: 'wait',   close: 'Waitlist' },
  { id: 't3', event: 'obedience',  title: 'Obedience Trial · Mid-Atlantic',   date: 'Sat, Jun 7',   time: '08.00 walkthrough',  loc: 'Annapolis, MD',  spots: 9,  cap: 24, status: 'open',   close: '20 days' },
  { id: 't4', event: 'protection', title: 'Protection Trial · National Team', date: 'Sat, Jun 14',  time: 'Closed brief',       loc: 'Lexington, KY',  spots: 16, cap: 16, status: 'closed', close: 'Closed' },
  { id: 't5', event: 'detection',  title: 'Summer Detection · Two-Day',       date: 'Jul 12–13',    time: '06.30 walkthrough',  loc: 'Burlington, VT', spots: 4,  cap: 32, status: 'open',   close: '6 weeks' },
  { id: 't6', event: 'tracking',   title: 'Tracking Trial · Coastal Sand',    date: 'Sat, Aug 2',   time: '04.30 first track',  loc: 'Cape May, NJ',   spots: 6,  cap: 18, status: 'open',   close: '11 weeks' },
];

function TrialRowInfo({ t }) {
  return (
    <div className="row-info">
      <div><Icon name="calendar" /> {t.date}</div>
      <div><Icon name="clock" /> {t.time}</div>
      <div><Icon name="map-pin" /> {t.loc}</div>
    </div>
  );
}

function TrialCard({ t, onOpen }) {
  return (
    <div className="trial-card" data-event={t.event} onClick={() => onOpen && onOpen(t)}>
      <div className="flex justify-between items-start gap-3">
        <div className="flex-col gap-1">
          <div className="eyebrow">{t.event[0].toUpperCase() + t.event.slice(1)} · Trial</div>
          <div className="h3" style={{ maxWidth: 460 }}>{t.title}</div>
        </div>
        {t.status === 'open' && <Pill variant="open">Open</Pill>}
        {t.status === 'wait' && <Pill variant="wait">Waitlist</Pill>}
        {t.status === 'closed' && <Pill variant="closed">Closed</Pill>}
      </div>
      <TrialRowInfo t={t} />
      <div className="footer-row">
        <div><b>{t.spots}</b> of {t.cap} spots {t.status === 'wait' ? 'on list' : 'filled'} · entries {t.close.includes('Closed') || t.close.includes('Waitlist') ? '' : 'close in'} {t.close}</div>
        <Button variant="primary" size="md" onClick={(e) => { e.stopPropagation(); onOpen && onOpen(t); }}>
          {t.status === 'closed' ? 'View brief' : 'Enter'}
        </Button>
      </div>
    </div>
  );
}

function TrialListSection({ title = 'Upcoming trials', limit, eyebrow = 'This Season' }) {
  const items = limit ? TRIALS.slice(0, limit) : TRIALS;
  return (
    <section className="section" data-screen-label="03 Trials List">
      <Container>
        <div className="flex-col gap-10">
          <div className="flex justify-between items-end gap-6 flex-wrap">
            <div className="flex-col gap-2" style={{ maxWidth: 640 }}>
              <Eyebrow>{eyebrow}</Eyebrow>
              <h2 className="h2">{title}</h2>
            </div>
            {limit && (
              <Button variant="plain" size="lg" onClick={() => window.kit.navigate('trials')}>
                Full schedule <Icon name="arrow-right" size={13} />
              </Button>
            )}
          </div>
          <div className="flex-col gap-3">
            {items.map(t => <TrialCard key={t.id} t={t} onOpen={(t) => window.kit.openTrial(t)} />)}
          </div>
        </div>
      </Container>
    </section>
  );
}

window.TrialListSection = TrialListSection;
window.TrialCard = TrialCard;
window.TRIALS = TRIALS;
