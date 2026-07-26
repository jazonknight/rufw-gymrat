/**
 * Workout Routine Templates UI Component
 * Handles rendering routine cards, duplicating routines, editing routine templates, and multi-set step builders.
 */

import { state, escapeHtml } from '../state.js';
import { copyWorkoutTemplate, deleteRoutine } from '../api.js';

export function renderRoutines() {
  const grid = document.getElementById('routines-grid');
  if (!grid) return;
  grid.innerHTML = '';

  if (!state.workoutRoutines || state.workoutRoutines.length === 0) {
    grid.innerHTML = '<p class="subtitle">No standalone workout routines created yet. Click "+ Create Workout Routine".</p>';
    return;
  }

  state.workoutRoutines.forEach(routine => {
    const card = document.createElement('div');
    card.className = 'exercise-card';

    let exSummary = '';
    if (routine.exercises) {
      Object.values(routine.exercises).forEach(we => {
        let setsSummary = (we.sets || []).map((s, i) => {
          return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight}lbs` : `${s.repCount} reps @ ${s.weight}lbs`}`;
        }).join(' | ');
        exSummary += `<div style="font-size: 0.85rem; margin-top: 4px;">• <strong>${escapeHtml(we.nameSnapshot || 'Exercise')} (${(we.sets || []).length} sets):</strong> ${setsSummary}</div>`;
      });
    }

    card.innerHTML = `
      <div class="card-top">
        <h3 class="ex-title">${escapeHtml(routine.name)}</h3>
        <span class="tag-version">Routine Template</span>
      </div>
      <p class="ex-desc">${escapeHtml(routine.description || 'Workout Routine Template')}</p>
      ${exSummary}
      <div class="card-actions" style="display: flex; gap: 8px; margin-top: 12px; flex-wrap: wrap;">
        <button class="btn btn-primary btn-sm btn-dup-routine" data-id="${routine.id}">📋 Duplicate Routine</button>
        <button class="btn btn-secondary btn-sm btn-edit-routine" data-id="${routine.id}">✏️ Edit Routine</button>
        <button class="btn btn-secondary btn-sm btn-del-routine" style="color: var(--accent-rose);" data-id="${routine.id}">🗑️ Delete</button>
      </div>
    `;

    card.querySelector('.btn-dup-routine')?.addEventListener('click', () => copyWorkoutTemplate(routine.id));
    card.querySelector('.btn-edit-routine')?.addEventListener('click', () => editRoutine(routine.id));
    card.querySelector('.btn-del-routine')?.addEventListener('click', () => deleteRoutine(routine.id));

    grid.appendChild(card);
  });
}

export function editRoutine(routineId) {
  const routine = state.workoutRoutines.find(r => r.id === routineId);
  if (!routine) return;

  document.getElementById('routine-id').value = routine.id;
  document.getElementById('routine-name').value = routine.name;
  document.getElementById('routine-desc').value = routine.description || '';
  document.getElementById('modal-routine-title').textContent = 'Edit Workout Routine';

  const builder = document.getElementById('routine-exercises-builder');
  if (builder) builder.innerHTML = '';

  if (routine.exercises && Object.keys(routine.exercises).length > 0) {
    Object.values(routine.exercises).forEach(we => {
      addRoutineExerciseStep(we);
    });
  } else {
    addRoutineExerciseStep();
  }

  document.getElementById('modal-routine')?.classList.add('active');
}

export function removeExerciseStep(btnEl) {
  const item = btnEl.closest('.routine-ex-builder-item');
  if (item) item.remove();
}

export function removeSetRow(btnEl) {
  const row = btnEl.closest('.routine-set-row');
  if (!row) return;
  const container = row.parentElement;
  row.remove();
  if (container) reindexSetRows(container.id);
}

export function reindexSetRows(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;
  const rows = container.querySelectorAll('.routine-set-row');
  rows.forEach((r, idx) => {
    const label = r.querySelector('.set-label');
    if (label) label.textContent = `Set ${idx + 1}`;
  });
}

export function addRoutineExerciseStep(existingWe = null) {
  const container = document.getElementById('routine-exercises-builder');
  if (!container) return;
  const index = container.children.length + 1;
  const exIdIndex = `ex_builder_${Date.now()}_${index}`;

  let optionsHtml = state.exerciseCatalog.map(ex => {
    const isSel = existingWe && existingWe.exerciseId === ex.id ? 'selected' : '';
    return `<option value="${ex.id}" ${isSel}>${escapeHtml(ex.name)} (v${ex.version})</option>`;
  }).join('');

  const div = document.createElement('div');
  div.className = 'routine-ex-builder-item';
  div.style.cssText = 'background: rgba(0,0,0,0.3); padding: 14px; margin-bottom: 14px; border-radius: 8px; border: 1px solid var(--border-color)';
  div.innerHTML = `
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
      <label style="font-size: 0.9rem; font-weight: 600;">Select Exercise from Library</label>
      <button type="button" class="btn btn-secondary btn-sm btn-remove-ex" style="padding: 4px 8px; color: var(--accent-rose);">🗑️ Remove Exercise</button>
    </div>
    <div class="form-group" style="margin-bottom: 10px;">
      <select class="r-ex-select">${optionsHtml || '<option value="">No Library Exercises Available</option>'}</select>
    </div>
    
    <div class="r-sets-container" id="sets_container_${exIdIndex}">
      <!-- Set rows dynamically added -->
    </div>

    <button type="button" class="btn btn-secondary btn-sm btn-add-set-row" style="margin-top: 6px;">+ Add Set</button>
  `;

  div.querySelector('.btn-remove-ex')?.addEventListener('click', (e) => removeExerciseStep(e.target));
  div.querySelector('.btn-add-set-row')?.addEventListener('click', () => addSetRowToRoutine(`sets_container_${exIdIndex}`));

  container.appendChild(div);

  if (existingWe && existingWe.sets && existingWe.sets.length > 0) {
    existingWe.sets.forEach(s => addSetRowToRoutine(`sets_container_${exIdIndex}`, s));
  } else {
    // Default to 1 initial set row for newly added exercise step
    addSetRowToRoutine(`sets_container_${exIdIndex}`);
  }
}

export function addSetRowToRoutine(containerId, setObj = null) {
  const container = document.getElementById(containerId);
  if (!container) return;

  const setIndex = container.children.length + 1;
  const setType = setObj ? (setObj.setType || 'reps') : 'reps';
  const valCount = setObj ? (setType === 'timed' ? (setObj.durationSeconds || 30) : (setObj.repCount || 10)) : 10;
  const weight = setObj ? (setObj.weight || 0) : 45;

  const setRow = document.createElement('div');
  setRow.className = 'routine-set-row';
  setRow.style.cssText = 'display: flex; gap: 10px; align-items: center; margin-bottom: 8px; background: rgba(255,255,255,0.03); padding: 8px; border-radius: 6px;';
  setRow.innerHTML = `
    <span class="set-label" style="font-size: 0.85rem; font-weight: 600; width: 45px;">Set ${setIndex}</span>
    <div style="flex: 1;">
      <select class="r-set-type" style="padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
        <option value="reps" ${setType === 'reps' ? 'selected' : ''}>Counting (Reps)</option>
        <option value="timed" ${setType === 'timed' ? 'selected' : ''}>Timed (Seconds)</option>
      </select>
    </div>
    <div style="flex: 1;">
      <input type="number" class="r-val" value="${valCount}" placeholder="Reps/Sec" style="width: 100%; padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
    </div>
    <div style="flex: 1;">
      <input type="number" class="r-weight" value="${weight}" placeholder="Weight (lbs)" style="width: 100%; padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
    </div>
    <button type="button" class="btn btn-secondary btn-sm btn-remove-set" style="padding: 4px 8px; color: var(--accent-rose);" title="Remove Set">🗑️</button>
  `;

  setRow.querySelector('.btn-remove-set')?.addEventListener('click', (e) => removeSetRow(e.target));

  container.appendChild(setRow);
  reindexSetRows(containerId);
}
