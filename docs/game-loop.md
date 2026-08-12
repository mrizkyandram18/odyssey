# Game Loop

## The Moment-to-Moment Loop

```
[See Invitation] → [Take a Turn] → [Contribute] → [See Result] → [Next Invitation]
```

1. **See Invitation.** The family's home screen presents one or more prompts:
   a daily turn, an in-progress quest, or a creative-space nudge.
2. **Take a Turn.** A family member opens the invitation. It is scoped to
   2–10 minutes of engagement.
3. **Contribute.** The member completes a small challenge or makes a choice.
   The contribution is written to the server immediately; the UI reflects it
   optimistically.
4. **See Result.** The world updates — a story scene advances, a Gift is
   granted, the family's progress ticks forward. Everyone's screen reflects
   this on the next reload.
5. **Next Invitation.** The loop naturally surfaces the next thing: a new
   daily turn, a handoff to another member, or the next course of a quest.

## The Daily Loop

```
Morning → Daily Turn → Micro-Mission / Creative Prompt
Evening → Family Check-in → Mission Progress / Story Reveal
```

- One **daily turn** unlocks each calendar day. It is intentionally light: a
  single micro-quest or a creative prompt.
- Completing the daily turn grants a small bundle of XP and advances at
  least one in-progress quest or unlocks a story beat.
- A **daily streak** multiplier (e.g., +10% XP for 3 days in a row) gently
  encourages consistency. Missing a day breaks the streak but causes no
  penalty to long-term progress.

## The Mission Loop

```
[Mission Offered] → [Family Reads Story] → [Exercises in Parallel/Sequence]
    → [All Exercises Done] → [Mission Resolved] → [Reward + Story Branch]
    → [New Missions / Journeys Unlocked]
```

- Missions are offered by the shared world (story-driven, not random).
- Exercises within a quest can be tackled in parallel by different family
  members or in a relay sequence.
- The family sees a shared quest board listing active missions, who's working on
  what, and what's needed next.
- Mission resolution triggers a story reveal and may unlock a new journey, creative
  space, or Journey Progress milestone.

## The Progression Loop

```
[Play] → [Experience & Collections] → [Level Up / Unlock] → [New Content] → [Play]
```

- Playing (completing turns and missions) generates XP, Collections, and Gifts.
- XP advances individual Explorer Levels and crew level.
- Gifts contain known Collections and cosmetics that feed the crew gallery and
  Inspirational tool unlocks.
- New content (realms, quest lines, creative tools) is gated behind story
  progress and Journey Progress thresholds to ensure a satisfying cadence.

## The Feedback Loop (Retention)

```
[Daily Turn] → [Small Win] → [Visible Progress] → [Motivation to Return]
```

- Every session ends with a visible marker of progress: a filled XP bar, a
  new Collection in the gallery, a story line revealed.
- The home screen always shows a clear "next thing" so there is no ambiguity
  about what to do.
- Gentle, opt-in notifications surface only when the family is likely to
  actually play (evening hours, weekends).

## The Journey Progress Loop

```
[Complete Missions in Journey] → [Journey Progress Fills]
    → [Journey Milestones + Journey Gift]
    → [Next Journey Unlocked]
    → [New Missions] → [Play]
```

- Missions within a journey advance the shared Journey Progress bar.
- Journey Progress milestones unlock new missions, a Journey Gift, and concept.
- When the bar is full, the next journey (and its missions) becomes available.

## Online Sync

The PWA fetches fresh world state from the server on foreground and after
network reconnect. There is no local write queue in the MVP — contributions
are sent immediately on submission. A fuller offline-first sync mechanism is
planned for a future phase (see [docs/future.md](future.md)).
