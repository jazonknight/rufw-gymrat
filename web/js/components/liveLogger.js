/**
 * Live Workout Logger & Rest Countdown Component
 * Manages live session elapsed timer, stateful set completion tracking, real-time per-set weight/rep input sync, set/exercise removal, and history submission.
 */

import { state, escapeHtml, API_BASE } from '../state.js';
import { toggleWorkoutExecuted, loadDashboardData } from '../api.js';

export function startLiveWorkout(planId, workoutId) {
  const plan = state.workoutPlans.find(p => p.id === planId);
  if (!plan) return;
  const workout = (plan.workouts || []).find(w => w.id === workoutId);
  if (!workout) return;

  // Clone workout object for live logging
  const clonedExercises = JSON.parse(JSON.stringify(workout.exercises || {}));
  
  // Ensure every set object has isCompleted state tracking
  Object.values(clonedExercises).forEach(we => {
    (we.sets || []).forEach(s => {
      if (typeof s.isCompleted === 'undefined') {
        s.isCompleted = false;
      }
    });
  });

  state.liveWorkoutSession = {
    planId: plan.id,
    workoutId: workout.id,
    workoutName: workout.name,
    exercises: clonedExercises
  };

  const titleEl = document.getElementById('live-workout-name');
  if (titleEl) titleEl.textContent = `Live: ${workout.name}`;

  const restEl = document.getElementById('live-rest-timer');
  if (restEl) restEl.textContent = '';
  
  // Reset and start live timer
  state.liveTimerSeconds = 0;
  clearInterval(state.liveTimerInterval);
  state.liveTimerInterval = setInterval(() => {
    state.liveTimerSeconds++;
    const hrs = String(Math.floor(state.liveTimerSeconds / 3600)).padStart(2, '0');
    const mins = String(Math.floor((state.liveTimerSeconds % 3600) / 60)).padStart(2, '0');
    const secs = String(state.liveTimerSeconds % 60).padStart(2, '0');
    const timerEl = document.getElementById('live-workout-timer');
    if (timerEl) timerEl.textContent = `⏱️ Session Duration: ${hrs}:${mins}:${secs}`;
  }, 1000);

  renderLiveWorkoutBody();
  document.getElementById('modal-live-workout')?.classList.add('active');
}

export function renderLiveWorkoutBody() {
  const body = document.getElementById('live-workout-body');
  if (!body) return;
  body.innerHTML = '';

  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises) return;

  Object.values(state.liveWorkoutSession.exercises).forEach(we => {
    const exDiv = document.createElement('div');
    exDiv.style.cssText = 'background: rgba(0,0,0,0.3); padding: 14px; margin-bottom: 14px; border-radius: 10px; border: 1px solid var(--border-color);';

    let setsHtml = (we.sets || []).map((s, idx) => {
      const isDone = s.isCompleted === true;
      const btnStyle = isDone ? 'background: var(--accent-emerald); color: #000;' : '';
      const btnText = isDone ? '✓ Completed' : '✓ Set Done';

      return `
        <div style="display: flex; gap: 8px; align-items: center; margin-top: 8px; font-size: 0.88rem; flex-wrap: wrap;">
          <span style="width: 48px; font-weight: 600;">Set ${idx + 1}:</span>
          <label>Weight (lbs):</label>
          <input type="number" class="live-weight live-weight-inp" data-ex="${we.exerciseId}" data-set="${idx}" value="${s.weight || 0}" 
                 style="width: 75px; padding: 6px; background: rgba(0,0,0,0.5); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
          
          <label>${s.setType === 'timed' ? 'Sec:' : 'Reps:'}</label>
          <input type="number" class="live-reps live-reps-inp" data-ex="${we.exerciseId}" data-set="${idx}" value="${s.setType === 'timed' ? (s.durationSeconds || 30) : (s.repCount || 10)}" 
                 style="width: 75px; padding: 6px; background: rgba(0,0,0,0.5); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
          
          <button type="button" class="btn btn-secondary btn-sm btn-toggle-set-done" style="${btnStyle}" data-ex="${we.exerciseId}" data-set="${idx}">${btnText}</button>
          <button type="button" class="btn btn-secondary btn-sm btn-remove-live-set" style="padding: 4px 8px; color: var(--accent-rose);" data-ex="${we.exerciseId}" data-set="${idx}" title="Remove Set">🗑️</button>
        </div>
      `;
    }).join('');

    exDiv.innerHTML = `
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
        <h4 style="color: var(--accent-emerald); font-size: 1.1rem;">${escapeHtml(we.nameSnapshot || 'Exercise')}</h4>
        <div style="display: flex; gap: 6px;">
          <button type="button" class="btn btn-secondary btn-sm btn-add-live-set" data-ex="${we.exerciseId}">+ Add Set</button>
          <button type="button" class="btn btn-secondary btn-sm btn-remove-live-ex" style="padding: 4px 8px; color: var(--accent-rose);" data-ex="${we.exerciseId}" title="Remove Exercise">🗑️ Remove Exercise</button>
        </div>
      </div>
      ${setsHtml}
    `;

    exDiv.querySelectorAll('.live-weight-inp').forEach(inp => {
      inp.addEventListener('input', (e) => updateLiveSetWeight(e.target.dataset.ex, parseInt(e.target.dataset.set), e.target.value));
    });

    exDiv.querySelectorAll('.live-reps-inp').forEach(inp => {
      inp.addEventListener('input', (e) => updateLiveSetReps(e.target.dataset.ex, parseInt(e.target.dataset.set), e.target.value));
    });

    exDiv.querySelectorAll('.btn-toggle-set-done').forEach(btn => {
      btn.addEventListener('click', (e) => toggleLiveSetCompleted(e.target.dataset.ex, parseInt(e.target.dataset.set)));
    });

    exDiv.querySelectorAll('.btn-remove-live-set').forEach(btn => {
      btn.addEventListener('click', (e) => removeLiveSet(e.target.dataset.ex, parseInt(e.target.dataset.set)));
    });

    exDiv.querySelector('.btn-add-live-set')?.addEventListener('click', (e) => addExtraLiveSet(e.target.dataset.ex));
    exDiv.querySelector('.btn-remove-live-ex')?.addEventListener('click', (e) => removeLiveExercise(e.target.dataset.ex));

    body.appendChild(exDiv);
  });
}

export function updateLiveSetWeight(exerciseId, setIdx, value) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;
  const sets = state.liveWorkoutSession.exercises[exerciseId].sets || [];
  if (sets[setIdx]) {
    sets[setIdx].weight = parseFloat(value) || 0;
  }
}

export function updateLiveSetReps(exerciseId, setIdx, value) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;
  const sets = state.liveWorkoutSession.exercises[exerciseId].sets || [];
  if (sets[setIdx]) {
    const val = parseFloat(value) || 0;
    if (sets[setIdx].setType === 'timed') {
      sets[setIdx].durationSeconds = val;
    } else {
      sets[setIdx].repCount = val;
    }
  }
}

export function toggleLiveSetCompleted(exerciseId, setIdx) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;
  const sets = state.liveWorkoutSession.exercises[exerciseId].sets || [];
  if (!sets[setIdx]) return;

  sets[setIdx].isCompleted = !sets[setIdx].isCompleted;
  if (sets[setIdx].isCompleted) {
    triggerRestTimer(30);
  }
  renderLiveWorkoutBody();
}

export function removeLiveSet(exerciseId, setIdx) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;
  const sets = state.liveWorkoutSession.exercises[exerciseId].sets || [];
  sets.splice(setIdx, 1);
  state.liveWorkoutSession.exercises[exerciseId].sets = sets;
  renderLiveWorkoutBody();
}

export function removeLiveExercise(exerciseId) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;
  delete state.liveWorkoutSession.exercises[exerciseId];
  renderLiveWorkoutBody();
}

export function addExtraLiveSet(exerciseId) {
  if (!state.liveWorkoutSession || !state.liveWorkoutSession.exercises[exerciseId]) return;

  const currentSets = state.liveWorkoutSession.exercises[exerciseId].sets || [];
  const lastSet = currentSets[currentSets.length - 1] || { setType: 'reps', repCount: 10, weight: 45 };

  currentSets.push({
    id: crypto.randomUUID(),
    setType: lastSet.setType,
    repCount: lastSet.repCount,
    durationSeconds: lastSet.durationSeconds,
    weight: lastSet.weight,
    perceivedEffort: 7,
    isCompleted: false
  });

  state.liveWorkoutSession.exercises[exerciseId].sets = currentSets;
  renderLiveWorkoutBody();
}

export function triggerRestTimer(seconds = 30) {
  let restRemaining = seconds;
  clearInterval(state.restTimerInterval);
  const timerEl = document.getElementById('live-rest-timer');

  state.restTimerInterval = setInterval(() => {
    restRemaining--;
    if (restRemaining <= 0) {
      clearInterval(state.restTimerInterval);
      if (timerEl) timerEl.textContent = '⚡ Rest Finished! Ready for next set.';
    } else {
      if (timerEl) timerEl.textContent = `⏱️ Rest Timer: ${restRemaining}s remaining`;
    }
  }, 1000);
}

export async function finishLiveWorkout() {
  if (!state.liveWorkoutSession) return;

  clearInterval(state.liveTimerInterval);
  clearInterval(state.restTimerInterval);

  let totalVolumeLbs = 0;

  // Calculate volume from completed / logged sets
  Object.values(state.liveWorkoutSession.exercises).forEach(we => {
    (we.sets || []).forEach(s => {
      const weight = s.weight || 0;
      if (s.setType !== 'timed') {
        const reps = s.repCount || 0;
        totalVolumeLbs += (weight * reps);
      }
    });
  });

  const historyLog = {
    id: crypto.randomUUID(),
    workoutId: state.liveWorkoutSession.workoutId,
    planId: state.liveWorkoutSession.planId,
    workoutName: state.liveWorkoutSession.workoutName,
    dateCompleted: new Date().toISOString(),
    durationSeconds: state.liveTimerSeconds,
    totalVolumeLbs: totalVolumeLbs,
    workoutSnapshot: {
      id: state.liveWorkoutSession.workoutId,
      name: state.liveWorkoutSession.workoutName,
      datePlanned: new Date().toISOString(),
      isExecuted: true,
      exercises: state.liveWorkoutSession.exercises
    }
  };

  await fetch(`${API_BASE}/history`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Session-ID': state.activeSessionId
    },
    body: JSON.stringify(historyLog)
  });

  await toggleWorkoutExecuted(state.liveWorkoutSession.planId, state.liveWorkoutSession.workoutId, true);

  document.getElementById('modal-live-workout')?.classList.remove('active');
  state.liveWorkoutSession = null;
  await loadDashboardData();
}
