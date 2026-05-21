/* eslint-disable */
// Training page — the front door to the paid product. Shows the 4 paths
// with imagery, lede, and a "Start path" CTA.

function PathRow({ ev, photoKind, headline, weeks }) {
  return (
    <div data-event={ev} className="grid grid-cols-2 gap-8 items-center" style={{ padding: '40px 0', borderBottom: '1px solid var(--mist-200)' }}>
      <div className="flex-col gap-3" style={{ maxWidth: 480 }}>
        <Eyebrow>{ev[0].toUpperCase() + ev.slice(1)} · Guided path</Eyebrow>
        <h3 className="h2" style={{ fontSize: '2.25rem' }}>{headline}</h3>
        <p className="body" style={{ marginTop: 4 }}>
          {weeks} weeks of structured sessions. Each block has a written drill, a reference video, and judge-style criteria. The path branches based on what you score.
        </p>
        <div className="flex gap-3" style={{ marginTop: 12 }}>
          <Button variant="event" size="lg">Start path</Button>
          <Button variant="plain" size="lg">Sample week <Icon name="arrow-right" /></Button>
        </div>
      </div>
      <PhotoPlaceholder kind={photoKind} caption={`${ev} · field study`} height="400px" maxHeight={400} eventTint={ev} />
    </div>
  );
}

function TrainingPage() {
  return (
    <main className="fade-up">
      <section className="section-lg" data-screen-label="05 Training">
        <Container>
          <div className="flex-col gap-16">
            <div className="flex-col gap-4" style={{ maxWidth: 720 }}>
              <Eyebrow>Training</Eyebrow>
              <h1 className="hero-h1">Four paths. One judged criterion at a time.</h1>
              <p className="lede" style={{ maxWidth: 580 }}>
                Each guided path is twelve weeks of drills, weekly video review, and judge-graded session logs.
              </p>
            </div>
            <div>
              <PathRow ev="obedience"  photoKind="ob"   headline="Heelwork that the judge will count." weeks={12} />
              <PathRow ev="tracking"   photoKind="track" headline="Read the wind. Read the ground."     weeks={12} />
              <PathRow ev="detection"  photoKind="det"   headline="Source. Verify. Indicate. Reward."   weeks={12} />
              <PathRow ev="protection" photoKind="prot"  headline="Drive that holds under pressure."    weeks={12} />
            </div>
          </div>
        </Container>
      </section>
    </main>
  );
}

window.TrainingPage = TrainingPage;
