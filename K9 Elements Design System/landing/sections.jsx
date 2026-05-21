// K9 Elements landing — sections

const Utility = () => (
  <div className="k9l-utility">
    <div className="stamps">
      <span>K9E v1 · Sanctioned Programme</span>
      <span>1,247 active members</span>
      <span>Spring season · open</span>
    </div>
    <div className="right">
      <a href="#">Find a Coach</a>
      <a href="#">Member Sign-In</a>
    </div>
  </div>
);

const Mast = () => (
  <header className="k9l-mast">
    <div className="mark">
      <span className="wm">K9 ELEMENTS</span>
      <span className="est">Est. MMXXVI · Working Dog Sport</span>
    </div>
    <nav>
      <a className="cur" href="#">Home</a>
      <a href="#">Disciplines</a>
      <a href="#">Trials</a>
      <a href="#">Officials</a>
      <a href="#">Rulebook</a>
    </nav>
    <div className="cta">
      <button className="k9l-action ghost">Sign In</button>
      <button className="k9l-action">Become a Member</button>
    </div>
  </header>
);

const Hero = ({ tweaks }) => {
  const angleStyle = { "--spine-deg": tweaks.heroAngle + "deg", "--spine-x": tweaks.heroOffset + "%" };
  return (
    <section className="k9l-hero" style={angleStyle}>
      <div className="k9l-spine" aria-hidden="true"></div>
      <div className="k9l-hero-grid">
        <div>
          <div className="k9l-hero-eyebrow">
            <span className="tag">§00 / Front Matter</span>
            <span>A working-dog sport — built deliberately</span>
            <span className="rule"></span>
          </div>
          <h1 className="k9l-hero-title">
            A new<br/>
            standard for<br/>
            <span className="accent">working&nbsp;dog</span><br/>
            <span>sport.</span>
          </h1>
          <p className="k9l-hero-lede">
            Four disciplines, one rulebook, one platform. K9 Elements is for handlers who would rather train under judges than algorithms — and for the dogs that have been waiting for the work.
          </p>
          <div className="k9l-hero-cta">
            <button className="k9l-action lg">Become a Member →</button>
            <span className="meta">$95 / year · cancel anytime · no dog required to start</span>
          </div>
        </div>
        <figure className="k9l-hero-frame">
          <div className="corner-mark"><span className="num">F-04</span> · LONGLINE RECALL · 06:42 CT</div>
          <div className="ph">
            <div className="big">FIELD 04</div>
            <div>WIDE FRAME · DOG &amp; HANDLER · 50YDS<br/>OBSERVED, NOT STAGED</div>
          </div>
          <div className="caption">
            <div>
              <div className="lab">Frisco, Texas</div>
              <div>North Texas Spring Trial · Day 02</div>
            </div>
            <div style={{textAlign:"right"}}>
              <div className="lab">Frame</div>
              <div>04 / 312</div>
            </div>
          </div>
        </figure>
      </div>
      <div className="k9l-hero-stats">
        <div>
          <div className="num">04</div>
          <div className="lab">Disciplines<br/>Obedience · Protection · Tracking · Detection</div>
        </div>
        <div>
          <div className="num">14</div>
          <div className="lab">Sanctioned trials<br/>in the next 60 days</div>
        </div>
        <div>
          <div className="num">312</div>
          <div className="lab">Open registrations<br/>spring season</div>
        </div>
        <div>
          <div className="num">1,247</div>
          <div className="lab">Active handlers<br/>across 38 affiliated clubs</div>
        </div>
      </div>
    </section>
  );
};

const Disciplines = ({ layout = "table" }) => (
  <section className="k9l-section cream">
    <div className="k9l-sechead">
      <div className="sigil">§01<span className="ref">The Disciplines</span></div>
      <div>
        <h2>Four disciplines. One ladder.</h2>
        <p className="dek">
          Every K9 Elements title is earned in one of four disciplines, scored against a published standard, and signed by a certified judge of record. No category bloat, no participation titles, no graduation through subscription length.
        </p>
      </div>
    </div>
    {layout === "table" && (
      <div className="k9l-disc-table">
        {DISCIPLINES.map((d) => (
          <div key={d.code} className={`k9l-disc-row ${d.code}`}>
            <div className="num-cell">
              <span className="swatch"></span>
              <span className="n">{d.n}</span>
            </div>
            <div className="name">{d.name}<span className="code">{d.short} · L1 — L3</span></div>
            <div className="desc">{d.desc}</div>
            <div className="levels">
              {d.levels.map((l) => <div key={l}><span className="l">{l.split(" / ")[0]}</span>{l.split(" / ")[1]}</div>)}
            </div>
            <div className="read">Standard</div>
          </div>
        ))}
      </div>
    )}
    {layout === "cards" && (
      <div className="k9l-disc-cards">
        {DISCIPLINES.map((d) => (
          <article key={d.code} className={`k9l-disc-card ${d.code}`}>
            <div className="head">
              <span className="n">{d.n}</span>
              <span className="code">{d.short}</span>
            </div>
            <h3>{d.name}</h3>
            <p>{d.desc}</p>
            <ul>
              {d.levels.map((l) => <li key={l}>{l}</li>)}
            </ul>
            <span className="read">Read the standard →</span>
          </article>
        ))}
      </div>
    )}
    {layout === "index" && (
      <div className="k9l-disc-index">
        {DISCIPLINES.map((d) => (
          <div key={d.code} className={`k9l-disc-idx ${d.code}`}>
            <div className="lead">
              <span className="n">{d.n}</span>
              <span className="title">{d.name}</span>
              <span className="dots"></span>
              <span className="code">{d.short}</span>
            </div>
            <p>{d.desc}</p>
          </div>
        ))}
      </div>
    )}
  </section>
);

const Flow = () => (
  <section className="k9l-section dark">
    <div className="k9l-sechead">
      <div className="sigil">§02<span className="ref">Progression</span></div>
      <div>
        <h2>Handler. Trial. Title.</h2>
        <p className="dek">
          The path is short, deliberate, and the same for everyone. There is no fast lane. There is no slow lane. There is the lane.
        </p>
      </div>
    </div>
    <div className="k9l-flow">
      {FLOW.map((s) => (
        <article key={s.n}>
          <div className="step">
            <span className="n">{s.n}</span>
            {s.step}
          </div>
          <h3>{s.title}</h3>
          <p>{s.body}</p>
          <ul>
            {s.rows.map(([k, v]) => <li key={k}><span className="k">{k}</span><span>{v}</span></li>)}
          </ul>
        </article>
      ))}
    </div>
  </section>
);

const Ticker = () => (
  <section className="k9l-section bone">
    <div className="k9l-sechead">
      <div className="sigil">§03<span className="ref">Live Activity</span></div>
      <div>
        <h2>
          <span className="k9l-accent" style={{background:"var(--success)", color:"transparent", WebkitBackgroundClip:"text", backgroundClip:"text", display:"inline-block", width:14, height:14, borderRadius:"50%", verticalAlign:"middle", marginRight:18, animation:"k9l-pulse 1.6s ease infinite"}}></span>
          On the field, right now.
        </h2>
        <p className="dek">
          Every qualifying run, registration, and title is broadcast as it happens. Names are public; scores are public; the rulebook reference for every deduction is public. Last seven entries from today's docket — North Texas Spring Trial, runs 1–7.
        </p>
      </div>
    </div>
    <div className="k9l-ticker">
      <div className="k9l-ticker-head">
        <div>Time</div>
        <div>Handler · Club</div>
        <div className="what">Action · Trial</div>
        <div className="where">Reference</div>
        <div style={{textAlign:"right", paddingRight:8}}>Score</div>
        <div></div>
      </div>
      {TICKER.map((r, i) => (
        <div key={i} className={`k9l-tick-row ${r.disc}`}>
          <div className="ts">{r.ts} CT</div>
          <div className="who">{r.who}<small>{r.region}</small></div>
          <div className="what">
            {i === 0 && <span className="live-dot"></span>}
            {r.what} · <span style={{color:"var(--muted)"}}>{r.where}</span>
          </div>
          <div className="where">§{(7 + i * 2)}.{(i + 1) % 9}{(i + 3) % 9}</div>
          <div className={`score ${r.q === true ? "q" : r.q === false ? "nq" : ""}`}>
            {r.score}{r.max && <small>/{r.max}</small>}
          </div>
          <div className="badge">{r.code.split("-")[1] || r.code}</div>
        </div>
      ))}
    </div>
  </section>
);

const Voice = () => (
  <section className="k9l-section cream">
    <div className="k9l-sechead">
      <div className="sigil">§04<span className="ref">Voice</span></div>
      <div>
        <h2>From the field.</h2>
      </div>
    </div>
    <div className="k9l-voice">
      <div>
        <blockquote>
          <span>“We&rsquo;ve been </span>
          <span className="accent">waiting twenty years</span>
          <span> for a sport that takes the work this seriously. The standard is the standard. </span>
          <span className="accent">The score is the score.</span>
          <span>”</span>
        </blockquote>
        <div className="attr">
          <div className="pic">KA</div>
          <div>
            <div className="who">K. Anderson</div>
            <div className="role">Master Judge · Iron Pack K9 · 22 years working dogs</div>
          </div>
        </div>
      </div>
      <figure className="k9l-voice-img">
        <div className="corner"><span className="num">P-12</span> · PORTRAIT · INTERMISSION</div>
        <div className="ph">PORTRAIT · JUDGE OF RECORD<br/>3/4 PROFILE · NATURAL LIGHT</div>
        <div className="cap">Frisco TX · April 2026 · 35mm</div>
      </figure>
    </div>
  </section>
);

const FinalCTA = () => (
  <section className="k9l-finalcta">
    <div className="k9l-finalcta-grid">
      <div>
        <h2>
          The work<br/>
          is the <span className="accent">work.</span>
        </h2>
        <p className="lede">
          You don&rsquo;t need a dog yet. You don&rsquo;t need a club. You need a rulebook, a coach within reach, and a season ahead. Membership opens all three.
        </p>
      </div>
      <div className="actions">
        <button className="k9l-action lg">Become a Member →</button>
        <button className="k9l-action ghost">Read the Rulebook</button>
        <span className="meta">$95 / year · cancel anytime</span>
      </div>
    </div>
  </section>
);

const Foot = () => (
  <footer className="k9l-foot">
    <div className="k9l-foot-grid">
      <div>
        <div className="wm">K9 ELEMENTS</div>
        <p className="lede">
          A working-dog sport &amp; sanctioning body. Operations across North America. Headquarters: Frisco, Texas. K9E is governed by an elected board of judges, decoys, and senior handlers.
        </p>
      </div>
      <div className="col">
        <h4>Sport</h4>
        <ul>
          <li><a href="#">Disciplines</a></li>
          <li><a href="#">Title Ladder</a></li>
          <li><a href="#">Trial Calendar</a></li>
          <li><a href="#">Rulebook v3.1</a></li>
          <li><a href="#">Standards Library</a></li>
        </ul>
      </div>
      <div className="col">
        <h4>Membership</h4>
        <ul>
          <li><a href="#">Become a Handler</a></li>
          <li><a href="#">Find a Coach</a></li>
          <li><a href="#">Affiliate a Club</a></li>
          <li><a href="#">Apply as Decoy</a></li>
          <li><a href="#">Apply as Judge</a></li>
        </ul>
      </div>
      <div className="col">
        <h4>Records</h4>
        <ul>
          <li><a href="#">Title Index</a></li>
          <li><a href="#">Regional Rankings</a></li>
          <li><a href="#">Trial Archive</a></li>
          <li><a href="#">Grievance Log</a></li>
        </ul>
      </div>
      <div className="col">
        <h4>About</h4>
        <ul>
          <li><a href="#">Mission</a></li>
          <li><a href="#">Board &amp; Officials</a></li>
          <li><a href="#">Press</a></li>
          <li><a href="#">Contact</a></li>
        </ul>
      </div>
    </div>
    <div className="k9l-foot-bottom">
      <div>© 2026 K9 Elements Sanctioning Body</div>
      <div className="ref">RULEBOOK §1.1 · v3.1 · APR 2026</div>
      <div>Frisco, TX · Bend, OR · Toronto, ON</div>
    </div>
  </footer>
);

Object.assign(window, { Utility, Mast, Hero, Disciplines, Flow, Ticker, Voice, FinalCTA, Foot });
