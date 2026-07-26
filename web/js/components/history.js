/**
 * Completed Workout History Log UI Component
 * Renders completed session duration, date completed, total volume (lbs) lifted, and exercise details.
 */

import { state, escapeHtml } from '../state.js';

export function renderHistory() {
  const container = document.getElementById('history-container');
  if (!container) return;
  container.innerHTML = '';

  if (!state.historyLogs || state.historyLogs.length === 0) {
    container.innerHTML = '<p class="subtitle">No completed workout history logs recorded yet. Complete a workout to start building history!</p>';
    return;
  }

  const sortedLogs = [...state.historyLogs].sort((a, b) => new Date(b.dateCompleted) - new Date(a.dateCompleted));

  sortedLogs.forEach(log => {
    const card = document.createElement('div');
    card.className = 'plan-card';

    const durationMin = Math.round((log.durationSeconds || 0) / 60);
    const dateStr = log.dateCompleted ? new Date(log.dateCompleted).toLocaleString() : 'Recent';

    let exDetails = '';
    if (log.workoutSnapshot && log.workoutSnapshot.exercises) {
      Object.values(log.workoutSnapshot.exercises).forEach(we => {
        let setPills = (we.sets || []).map((s, i) => {
          return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight} lbs` : `${s.repCount} reps @ ${s.weight} lbs`}`;
        }).join(' | ');
        exDetails += `<div><strong>${escapeHtml(we.nameSnapshot || 'Exercise')}:</strong> ${setPills}</div>`;
      });
    }

    card.innerHTML = `
      <div class="plan-header">
        <div>
          <h3>🏋️‍♂️ ${escapeHtml(log.workoutName || 'Completed Workout')}</h3>
          <p class="subtitle">Completed: ${dateStr} • Duration: ${durationMin} mins</p>
        </div>
        <span class="badge" style="background: var(--gradient-accent); color: #000;">${(log.totalVolumeLbs || 0).toLocaleString()} lbs Volume</span>
      </div>
      <div style="font-size: 0.9rem; line-height: 1.6;">
        ${exDetails}
      </div>
    `;
    container.appendChild(card);
  });
}
