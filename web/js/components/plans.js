/**
 * Training Plans UI Component
 * Handles training program list rendering, scheduled date instances, duplication, and plan creation.
 */

import { state, escapeHtml } from '../state.js';
import { copyPlan, toggleWorkoutExecuted } from '../api.js';
import { startLiveWorkout } from './liveLogger.js';

export function renderPlans() {
  const container = document.getElementById('plans-container');
  if (!container) return;
  container.innerHTML = '';

  if (!state.workoutPlans || state.workoutPlans.length === 0) {
    container.innerHTML = '<p class="subtitle">No training plans created yet. Click "+ Create Training Plan".</p>';
    return;
  }

  state.workoutPlans.forEach(plan => {
    const card = document.createElement('div');
    card.className = 'plan-card';

    let workoutsHtml = '';
    if (plan.workouts && plan.workouts.length > 0) {
      const sortedWorkouts = [...plan.workouts].sort((a, b) => new Date(a.datePlanned) - new Date(b.datePlanned));

      workoutsHtml = sortedWorkouts.map(w => {
        let exHtml = '';
        if (w.exercises) {
          Object.values(w.exercises).forEach(we => {
            let setsStr = we.sets ? we.sets.map((s, i) => {
              return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight}lbs` : `${s.repCount} reps @ ${s.weight}lbs`}`;
            }).join(' | ') : '';

            exHtml += `<div><strong>${escapeHtml(we.nameSnapshot || 'Exercise')}:</strong> ${setsStr}</div>`;
          });
        }

        const dateStr = w.datePlanned ? new Date(w.datePlanned).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' }) : 'Unscheduled';

        return `
          <div class="workout-item ${w.isExecuted ? 'executed' : ''}">
            <div class="workout-item-header">
              <div>
                <h4 style="font-size: 1.1rem; color: var(--text-main);">${escapeHtml(w.name)}</h4>
                <span class="scheduled-date">📅 Planned: ${dateStr}</span>
              </div>
              <div style="display: flex; gap: 8px; align-items: center;">
                <button class="btn btn-accent btn-sm btn-start-live" data-plan="${plan.id}" data-workout="${w.id}">
                  ▶ Start Workout
                </button>
                <button class="btn btn-secondary btn-sm btn-toggle-exec" data-plan="${plan.id}" data-workout="${w.id}" data-exec="${!w.isExecuted}">
                  ${w.isExecuted ? '✓ Executed' : 'Mark Executed'}
                </button>
              </div>
            </div>
            <p class="ex-desc">${escapeHtml(w.description || '')}</p>
            ${exHtml}
          </div>
        `;
      }).join('');
    }

    card.innerHTML = `
      <div class="plan-header">
        <div>
          <h3>${escapeHtml(plan.name)}</h3>
          <p class="subtitle">${escapeHtml(plan.description || '')} • ${plan.workouts ? plan.workouts.length : 0} Scheduled Sessions</p>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
          <button class="btn btn-secondary btn-sm btn-copy-plan" data-id="${plan.id}">📋 Duplicate Plan</button>
          <span class="badge">${plan.status || 'Active'}</span>
        </div>
      </div>
      <div class="workouts-subgrid">
        ${workoutsHtml || '<p class="subtitle">No workout instances scheduled in this plan.</p>'}
      </div>
    `;

    card.querySelector('.btn-copy-plan')?.addEventListener('click', () => copyPlan(plan.id));

    card.querySelectorAll('.btn-start-live').forEach(btn => {
      btn.addEventListener('click', () => startLiveWorkout(btn.dataset.plan, btn.dataset.workout));
    });

    card.querySelectorAll('.btn-toggle-exec').forEach(btn => {
      btn.addEventListener('click', () => toggleWorkoutExecuted(btn.dataset.plan, btn.dataset.workout, btn.dataset.exec === 'true'));
    });

    container.appendChild(card);
  });
}

export function addPlanScheduleStep() {
  const container = document.getElementById('plan-schedule-builder');
  if (!container) return;
  const index = container.children.length + 1;
  const todayStr = new Date().toISOString().split('T')[0];

  let routinesHtml = state.workoutRoutines.map(r => `<option value="${r.id}">${escapeHtml(r.name)}</option>`).join('');

  const div = document.createElement('div');
  div.className = 'plan-schedule-row';
  div.style.cssText = 'background: rgba(0,0,0,0.3); padding: 12px; margin-bottom: 12px; border-radius: 8px; border: 1px solid var(--border-color)';
  div.innerHTML = `
    <div style="display: flex; gap: 12px; align-items: flex-end;">
      <div class="form-group" style="flex: 2; margin-bottom: 0;">
        <label>Session #${index} Routine Template</label>
        <select class="p-routine-select">${routinesHtml || '<option value="">No Routine Templates Available</option>'}</select>
      </div>
      <div class="form-group" style="flex: 1; margin-bottom: 0;">
        <label>Scheduled Date</label>
        <input type="date" class="p-date-input" value="${todayStr}">
      </div>
    </div>
  `;
  container.appendChild(div);
}
