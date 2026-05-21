/* eslint-disable */
// Hero — image-forward, copy-light. Headline, lede, two CTAs, full-bleed video tile.

function Hero() {
  return (
    <section className="section-lg" data-screen-label="01 Home Hero">
      <Container>
        <div className="flex-col gap-16">
          <div className="flex-col items-start gap-6 fade-up" style={{ maxWidth: 920 }}>
            <Eyebrow>Spring 2026 schedule is live</Eyebrow>
            <h1 className="hero-h1">Built for the four events that matter.</h1>
            <p className="lede" style={{ maxWidth: 620 }}>
              Trial dates, guided training paths, and the people who organize the sport — in one place.
            </p>
            <div className="flex gap-4" style={{ marginTop: 8 }}>
              <Button variant="primary" size="lg" onClick={() => window.kit.navigate('trials')}>See the schedule</Button>
              <Button variant="plain" size="lg" onClick={() => window.kit.navigate('training')}>
                Explore training <Icon name="arrow-right" size={13} />
              </Button>
            </div>
          </div>
          <PhotoPlaceholder kind="field" caption="Spring open · Reston · Detection final" height="56vw" maxHeight={620} />
        </div>
      </Container>
    </section>
  );
}

window.Hero = Hero;
