/* eslint-disable */
// PricingTiers — three-tier with feature lists, used on /pricing.

const TIERS = [
  { name: 'Spectator',  price: 'Free',   per: '',           desc: 'Public trial schedule, results, and brief excerpts.', features: [
    'Full trial schedule across the four events',
    'Public results & rankings',
    'Find a club',
    'Read excerpts of training paths',
  ], cta: 'Create account', variant: 'soft' },
  { name: 'Pathfinder', price: '$24',    per: '/month',     desc: 'Full access to the four guided event paths, plus unlimited trial entries.', features: [
    '12-week paths · Obedience, Tracking, Detection, Protection',
    'Unlimited trial entries & waitlist priority',
    'Judge-graded video review (4 / month)',
    'Training library — 200+ drills',
    'Trial briefs delivered the night before',
  ], cta: 'Start free trial', badge: 'Most popular', variant: 'primary' },
  { name: 'Club',       price: '$199',   per: '/month',     desc: 'Run trials, manage rosters, and grade entries on the K9 Elements platform.', features: [
    'Everything in Pathfinder for 5 seats',
    'Host trials on the platform',
    'Roster &amp; payment management',
    'Judging scorecards on tablet',
    'Live event leaderboard',
  ], cta: 'Talk to us', variant: 'soft' },
];

function Tier({ t }) {
  return (
    <div className="tier">
      <div className="flex-col gap-1">
        <div className="flex items-center justify-between">
          {t.badge && <span className="pill" style={{ background: 'rgb(from var(--mist-950) r g b / 0.10)', color: 'var(--mist-950)', textTransform: 'none', letterSpacing: 0, fontWeight: 500, fontSize: 12 }}>{t.badge}</span>}
          <h3 className="h3" style={{ fontSize: 24 }}>{t.name}</h3>
        </div>
        <div className="price"><span className="amt">{t.price}</span>{t.per && <span className="per">{t.per}</span>}</div>
        <p className="body-sm" style={{ marginTop: 14 }}>{t.desc}</p>
        <ul style={{ marginTop: 14 }}>
          {t.features.map((f, i) => (
            <li key={i}>
              <Icon name="checkmark" size={13} style={{ color: 'var(--mist-950)' }} />
              <span dangerouslySetInnerHTML={{ __html: f }} />
            </li>
          ))}
        </ul>
      </div>
      <Button variant={t.variant} size="lg" className="" >{t.cta}</Button>
    </div>
  );
}

function PricingTiers() {
  return (
    <section className="section-lg" data-screen-label="06 Pricing">
      <Container>
        <div className="flex-col gap-10">
          <div className="flex-col gap-6" style={{ maxWidth: 640 }}>
            <Eyebrow>Membership</Eyebrow>
            <h2 className="h2">Free to spectate. Pay only when you train or host.</h2>
            <p className="lede" style={{ maxWidth: 560 }}>One subscription covers all four events. Cancel any time — the schedule stays public.</p>
          </div>
          <div className="grid grid-cols-3 gap-2">
            {TIERS.map(t => <Tier key={t.name} t={t} />)}
          </div>
        </div>
      </Container>
    </section>
  );
}

window.PricingTiers = PricingTiers;
