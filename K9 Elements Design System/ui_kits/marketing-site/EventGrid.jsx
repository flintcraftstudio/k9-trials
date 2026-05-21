/* eslint-disable */
// EventGrid — 4-event feature block. Each tile carries data-event so the
// eyebrow color and 3px rail come from the design system, not a one-off prop.

const EVENTS = [
  { id: 'obedience',  label: 'Obedience',  headline: 'Precision under load.',     blurb: 'Heelwork, recalls, retrieves, and the long down — graded by the criteria judges actually score on.' },
  { id: 'protection', label: 'Protection', headline: 'Controlled aggression.',    blurb: 'Bitework progression, courage testing, handler defense. Built around a serious training culture.' },
  { id: 'tracking',   label: 'Tracking',   headline: 'Read the ground.',          blurb: 'Article work, scent discrimination, weather-realistic terrain. The quiet event.' },
  { id: 'detection',  label: 'Detection',  headline: 'Find the source.',          blurb: 'Single-odor and multi-odor sets across four target areas. Sport-specific source rewards.' },
];

function EventGrid() {
  return (
    <section className="section" data-screen-label="02 Events Grid">
      <Container>
        <div className="flex-col gap-10">
          <div className="flex-col gap-6" style={{ maxWidth: 640 }}>
            <Eyebrow>The Four Events</Eyebrow>
            <h2 className="h2">One sport, four disciplines — each with its own path and its own judge.</h2>
          </div>
          <div className="grid grid-cols-2 gap-3">
            {EVENTS.map(e => (
              <a className="event-tile" data-event={e.id} key={e.id} href={`#training-${e.id}`}
                 onClick={(ev) => { ev.preventDefault(); window.kit.navigate('training'); }}>
                <div className="rail" />
                <div className="body">
                  <div className="eyebrow">{e.label}</div>
                  <div className="h3">{e.headline}</div>
                  <p className="body-sm" style={{ marginTop: 4 }}>{e.blurb}</p>
                  <div style={{ marginTop: 8 }}>
                    <span className="body-sm" style={{ color: 'var(--mist-950)', fontWeight: 500 }}>
                      Open the path <Icon name="arrow-right" size={13} style={{ marginLeft: 4 }} />
                    </span>
                  </div>
                </div>
              </a>
            ))}
          </div>
        </div>
      </Container>
    </section>
  );
}

window.EventGrid = EventGrid;
window.K9_EVENTS = EVENTS;
