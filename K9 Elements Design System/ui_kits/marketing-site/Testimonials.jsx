/* eslint-disable */
// Testimonial block — three-column grid.

const QUOTES = [
  { q: 'The tracking path got my green dog to a TD title in twelve weeks. The drills are written by judges, and you can feel it in how the criteria scale.', name: 'Petra Vogel',   by: 'Decoy & handler · Region 4', tint: 'tracking'  },
  { q: 'Entry was painless — I picked the trial, paid for the slot, and showed up. The brief was in my inbox the night before.',                              name: 'Marcus Lee',    by: 'Handler · Mid-Atlantic',     tint: 'detection' },
  { q: 'Watching the same judging criteria in the video reviews and at the trial — that\u2019s the part that makes this feel like a real sport.',             name: 'Dana Korhonen', by: 'Decoy & club president',     tint: 'obedience' },
];

function Testimonials() {
  return (
    <section className="section" data-screen-label="04 Testimonials">
      <Container>
        <div className="flex-col gap-10">
          <div className="flex-col gap-2" style={{ maxWidth: 640 }}>
            <Eyebrow>From the field</Eyebrow>
            <h2 className="h2">Built by handlers and judges, for handlers and judges.</h2>
          </div>
          <div className="grid grid-cols-3 gap-3">
            {QUOTES.map((q, i) => (
              <figure key={i} className="quote-card">
                <blockquote>{q.q}</blockquote>
                <figcaption className="flex items-center gap-3">
                  <div style={{ width: 42, height: 42, borderRadius: 9999, background: `var(--${q.tint}-300)`, outline: '1px solid rgb(0 0 0 / 0.05)' }} />
                  <div>
                    <div style={{ fontWeight: 600, fontSize: 14, color: 'var(--mist-950)' }}>{q.name}</div>
                    <div className="body-sm">{q.by}</div>
                  </div>
                </figcaption>
              </figure>
            ))}
          </div>
        </div>
      </Container>
    </section>
  );
}

window.Testimonials = Testimonials;
