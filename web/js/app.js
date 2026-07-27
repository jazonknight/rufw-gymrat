/**
 * GymRat SPA Main Application Entry Point
 * Initializes navigation tabs, modals, file upload dropzone, version badge, and session bootstrap.
 */

import { state, APP_VERSION, API_BASE, updateSessionUI } from './state.js';
import { createNewSession, loadDashboardData, fetchExercises, fetchWorkouts, fetchPlans, exportVault } from './api.js';
import { initFilters, renderCatalog } from './components/catalog.js';
import { addRoutineExerciseStep } from './components/routines.js';
import { addPlanScheduleStep } from './components/plans.js';
import { finishLiveWorkout } from './components/liveLogger.js';

document.addEventListener('DOMContentLoaded', async () => {
  const versionBadge = document.getElementById('app-version-badge');
  if (versionBadge) versionBadge.textContent = APP_VERSION;

  initTabs();
  initModals();
  initUpload();
  initFilters();

  if (!state.activeSessionId) {
    await createNewSession();
  } else {
    updateSessionUI(state.activeSessionId);
    await loadDashboardData();
  }

  document.getElementById('btn-new-session')?.addEventListener('click', createNewSession);
  document.getElementById('btn-export-vault')?.addEventListener('click', exportVault);
  document.getElementById('toggle-removed')?.addEventListener('change', renderCatalog);
});

// Navigation Tabs Setup
function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => t.classList.remove('active'));
      document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

      tab.classList.add('active');
      const targetId = `tab-${tab.dataset.tab}`;
      const targetEl = document.getElementById(targetId);
      if (targetEl) targetEl.classList.add('active');
    });
  });
}

// Modals & Form Submission Setup
function initModals() {
  const modalEx = document.getElementById('modal-exercise');
  const modalRoutine = document.getElementById('modal-routine');
  const modalPlan = document.getElementById('modal-plan');
  const modalLive = document.getElementById('modal-live-workout');

  // Open Exercise Modal
  document.getElementById('btn-add-exercise')?.addEventListener('click', () => {
    document.getElementById('ex-id').value = '';
    document.getElementById('modal-exercise-title').textContent = 'New Exercise';
    document.getElementById('form-exercise').reset();
    modalEx?.classList.add('active');
  });

  // Open Routine Modal
  document.getElementById('btn-add-routine')?.addEventListener('click', () => {
    document.getElementById('form-routine').reset();
    document.getElementById('routine-id').value = '';
    document.getElementById('modal-routine-title').textContent = 'Create Workout Routine';
    const builder = document.getElementById('routine-exercises-builder');
    if (builder) builder.innerHTML = '';
    addRoutineExerciseStep();
    modalRoutine?.classList.add('active');
  });

  // Open Plan Modal
  document.getElementById('btn-add-plan')?.addEventListener('click', () => {
    document.getElementById('form-plan').reset();
    const builder = document.getElementById('plan-schedule-builder');
    if (builder) builder.innerHTML = '';
    addPlanScheduleStep();
    modalPlan?.classList.add('active');
  });

  document.getElementById('btn-finish-live-workout')?.addEventListener('click', finishLiveWorkout);

  // Close modal buttons
  document.querySelectorAll('.modal-close, .modal-cancel').forEach(btn => {
    btn.addEventListener('click', () => {
      modalEx?.classList.remove('active');
      modalRoutine?.classList.remove('active');
      modalPlan?.classList.remove('active');
      modalLive?.classList.remove('active');
      clearInterval(state.liveTimerInterval);
      clearInterval(state.restTimerInterval);
    });
  });

  // Exercise Form Submit
  document.getElementById('form-exercise')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('ex-id').value;
    const name = document.getElementById('ex-name').value;
    const category = document.getElementById('ex-category').value;
    const description = document.getElementById('ex-desc').value;

    const method = id ? 'PUT' : 'POST';
    await fetch(`${API_BASE}/exercises`, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify({ id, name, category, description })
    });

    modalEx?.classList.remove('active');
    await fetchExercises();
  });

  document.getElementById('btn-add-routine-ex-step')?.addEventListener('click', () => addRoutineExerciseStep());
  document.getElementById('btn-add-plan-schedule-step')?.addEventListener('click', () => addPlanScheduleStep());

  // Routine Form Submit
  document.getElementById('form-routine')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const routineId = document.getElementById('routine-id').value;
    const name = document.getElementById('routine-name').value;
    const description = document.getElementById('routine-desc').value;

    const exEls = document.querySelectorAll('.routine-ex-builder-item');
    const exercisesMap = {};

    exEls.forEach(el => {
      const exSelect = el.querySelector('.r-ex-select');
      const selectedExId = exSelect ? exSelect.value : '';
      const selectedEx = state.exerciseCatalog.find(e => e.id === selectedExId) || { name: 'Custom Exercise', version: 1 };

      const setRows = el.querySelectorAll('.routine-set-row');
      const setsArr = [];

      setRows.forEach(sRow => {
        const setType = sRow.querySelector('.r-set-type').value;
        const valCount = parseInt(sRow.querySelector('.r-val').value) || 10;
        const weight = parseFloat(sRow.querySelector('.r-weight').value) || 0;

        setsArr.push({
          id: crypto.randomUUID(),
          setType: setType,
          repCount: setType === 'reps' ? valCount : 0,
          durationSeconds: setType === 'timed' ? valCount : 0,
          weight: weight,
          perceivedEffort: 7
        });
      });

      if (selectedExId) {
        exercisesMap[selectedExId] = {
          exerciseId: selectedExId,
          exerciseVersion: selectedEx.version,
          nameSnapshot: selectedEx.name,
          sets: setsArr
        };
      }
    });

    const targetRoutine = {
      id: routineId || crypto.randomUUID(),
      name,
      description,
      datePlanned: new Date().toISOString(),
      isExecuted: false,
      exercises: exercisesMap
    };

    const method = routineId ? 'PUT' : 'POST';
    await fetch(`${API_BASE}/workouts`, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify(targetRoutine)
    });

    modalRoutine?.classList.remove('active');
    await fetchWorkouts();
  });

  // Plan Form Submit
  document.getElementById('form-plan')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('plan-name').value;
    const description = document.getElementById('plan-desc').value;

    const rowEls = document.querySelectorAll('.plan-schedule-row');
    const selectedWorkouts = [];

    rowEls.forEach(row => {
      const rId = row.querySelector('.p-routine-select').value;
      const dateVal = row.querySelector('.p-date-input').value;
      const foundRoutine = state.workoutRoutines.find(r => r.id === rId);

      if (foundRoutine) {
        const plannedDate = dateVal ? new Date(dateVal + 'T00:00:00').toISOString() : new Date().toISOString();

        selectedWorkouts.push({
          id: crypto.randomUUID(),
          name: foundRoutine.name,
          description: foundRoutine.description,
          datePlanned: plannedDate,
          isExecuted: false,
          exercises: JSON.parse(JSON.stringify(foundRoutine.exercises || {}))
        });
      }
    });

    const newPlan = {
      id: crypto.randomUUID(),
      name,
      description,
      status: 'active',
      datePlanned: new Date().toISOString(),
      workouts: selectedWorkouts
    };

    await fetch(`${API_BASE}/plans`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify(newPlan)
    });

    modalPlan?.classList.remove('active');
    await fetchPlans();
  });
}

// File Upload Drop Zone Setup
function initUpload() {
  const dropZone = document.getElementById('drop-zone');
  const fileInput = document.getElementById('file-input');
  const statusMsg = document.getElementById('upload-status');
  const btnBrowse = document.getElementById('btn-browse-file');

  if (!dropZone || !fileInput) return;

  btnBrowse?.addEventListener('click', () => fileInput.click());

  fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
      handleFileUpload(e.target.files[0]);
    }
  });

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.style.borderColor = 'var(--accent-emerald)';
  });

  dropZone.addEventListener('dragleave', () => {
    dropZone.style.borderColor = 'var(--border-accent)';
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.style.borderColor = 'var(--border-accent)';
    if (e.dataTransfer.files.length > 0) {
      handleFileUpload(e.dataTransfer.files[0]);
    }
  });

  async function handleFileUpload(file) {
    if (!file.name.endsWith('.json')) {
      if (statusMsg) statusMsg.textContent = '❌ Invalid file type. Please upload a .json vault file.';
      return;
    }

    try {
      if (statusMsg) statusMsg.textContent = '⏳ Uploading vault file...';
      const res = await fetch(`${API_BASE}/vault/upload`, {
        method: 'POST',
        headers: {
          'X-Session-ID': state.activeSessionId
        },
        body: file
      });

      if (!res.ok) throw new Error('Upload failed');

      if (statusMsg) statusMsg.textContent = '✅ Vault data uploaded successfully!';
      await loadDashboardData();
    } catch (err) {
      if (statusMsg) statusMsg.textContent = `❌ Error: ${err.message}`;
    }
  }
}
