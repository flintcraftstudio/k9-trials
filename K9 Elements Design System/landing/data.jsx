// K9 Elements landing — content data

const DISCIPLINES = [
  {
    code: "ob",
    n: "01",
    name: "Obedience",
    short: "K9E-OB",
    desc: "Precision under distraction. Heeling, recalls, positions in motion, send-aways. The foundation discipline — every other discipline assumes a dog who will hold position and come when called.",
    levels: ["L1 / 18mo+", "L2 / req. K9E1", "L3 / req. K9E2"]
  },
  {
    code: "pr",
    n: "02",
    name: "Protection",
    short: "K9E-PR",
    desc: "Controlled bite work and guarding under judge command. Tests the dog's ability to engage on cue, hold, release on first command, and remain socially neutral when the work is done.",
    levels: ["L1 / req. K9E1-OB", "L2 / req. K9E1-PR", "L3 / req. K9E2-PR"]
  },
  {
    code: "tr",
    n: "03",
    name: "Tracking",
    short: "K9E-TR",
    desc: "Aged scent on natural ground. Articles laid by a stranger, recovered in sequence. Distance and age increase by level. Weather is part of the test, not an excuse.",
    levels: ["L1 / 30 min, ½ mi", "L2 / 90 min, 1 mi", "L3 / 3 hr, 2 mi"]
  },
  {
    code: "dt",
    n: "04",
    name: "Detection",
    short: "K9E-DT",
    desc: "Source-odor location in defined search areas. Vehicles, structures, open field. Indication on source, no false alerts, working through environmental contamination.",
    levels: ["L1 / 1 odor, 2 areas", "L2 / 2 odors, 3 areas", "L3 / 3 odors, 4 areas"]
  }
];

const FLOW = [
  {
    n: "01",
    step: "Step One · Become a Handler",
    title: "Join · Log · Begin",
    body: "Membership opens the door. You get a profile, a development log, access to the rulebook, and a coach-finder for your region. No dog required to start.",
    rows: [
      ["Cost", "$95 / year"],
      ["Includes", "Rulebook · Dev log · Coach finder"],
      ["Then", "Find a Level 1 trial"]
    ]
  },
  {
    n: "02",
    step: "Step Two · Enter a Trial",
    title: "Train · Enter · Run",
    body: "Sanctioned trials are listed by region and level. Register through the platform, upload vaccination records, e-sign the waiver. Your run is on the schedule the moment you pay.",
    rows: [
      ["Lead time", "60–90 days"],
      ["Format", "1–3 day sanctioned"],
      ["Recorded", "Yes · live-streamed"]
    ]
  },
  {
    n: "03",
    step: "Step Three · Earn a Title",
    title: "Qualify · Title · Progress",
    body: "Qualifying scores are issued same-day, signed by the judge of record. The title sheet enters your record, unlocks the next level, and posts to the public ranking sheet for your region.",
    rows: [
      ["Issued", "Same-day, on field"],
      ["Format", "K9E[level]-[disc]"],
      ["Public", "Visible in rankings"]
    ]
  }
];

const TICKER = [
  { ts: "11:42", who: "S. Martinez", region: "Iron Pack K9 — TX", what: "earned title", where: "North TX Spring", code: "K9E1-DT", score: "88", max: "100", q: true, disc: "dt" },
  { ts: "10:18", who: "R. Powell", region: "Lone Star K9 — TX", what: "qualifying run", where: "North TX Spring", code: "K9E2-OB", score: "82", max: "100", q: true, disc: "ob" },
  { ts: "09:55", who: "K. Lee", region: "Cascade Working — WA", what: "registered", where: "PNW Tracking · Jun 21", code: "K9E1-TR", score: "—", max: "", q: null, disc: "tr" },
  { ts: "09:31", who: "M. Torres", region: "Apprentice Decoy", what: "advanced level", where: "Decoy College", code: "PROV → CERT", score: "PASS", max: "", q: true, disc: "pr" },
  { ts: "09:02", who: "B. Choi", region: "Iron Pack K9 — TX", what: "non-qual run", where: "North TX Spring", code: "K9E1-PR", score: "62", max: "100", q: false, disc: "pr" },
  { ts: "08:47", who: "Iron Pack K9", region: "Affiliated Club", what: "submitted trial", where: "Fall Trial · Oct 3–5", code: "OB · PR · TR", score: "—", max: "", q: null, disc: "ob" },
  { ts: "08:14", who: "T. Nguyen", region: "Lone Star K9 — TX", what: "earned title", where: "North TX Spring", code: "K9E1-OB", score: "84", max: "100", q: true, disc: "ob" }
];

Object.assign(window, { DISCIPLINES, FLOW, TICKER });
