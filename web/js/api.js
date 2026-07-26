/**
 * GymRat REST API Client Module
 * Provides asynchronous network requests to the Go REST API backend.
 */

import { API_BASE, state, updateSessionUI } from './state.js';
import { renderCatalog, updateCategoryDropdownOptions } from './components/catalog.js';
import { renderRoutines } from './components/routines.js';
import { renderPlans } from './components/plans.js';
import { renderHistory } from './components/history.js';

/**
 * Creates a fresh workspace session on the server and loads initial data.
 */
export async function createNewSession() {
  try {
    const el = document.getElementById('active-session-id');
    if (el) el.textContent = 'Creating...';

    const res = await fetch(`${API_BASE}/session/new`, { method: 'POST' });
    const data = await res.json();
    state.activeSessionId = data.sessionId;
    localStorage.setItem('gymrat_session_id', state.activeSessionId);
    updateSessionUI(state.activeSessionId);
    await loadDashboardData();
  } catch (err) {
    console.error('Failed to create new session:', err);
    const el = document.getElementById('active-session-id');
    if (el) el.textContent = 'Error Connecting';
  }
}

/**
 * Loads all dashboard sections concurrently from the Go REST backend.
 */
export async function loadDashboardData() {
  await fetchExercises();
  await fetchWorkouts();
  await fetchPlans();
  await fetchHistory();
}

export async function fetchExercises() {
  try {
    const res = await fetch(`${API_BASE}/exercises`, {
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    if (!res.ok) {
      console.warn('Invalid session response, initializing new session...');
      await createNewSession();
      return;
    }
    const data = await res.json();
    state.exerciseCatalog = Array.isArray(data) ? data : [];
    updateCategoryDropdownOptions();
    renderCatalog();
  } catch (err) {
    console.error('Error fetching exercises:', err);
    state.exerciseCatalog = [];
  }
}

export async function fetchWorkouts() {
  try {
    const res = await fetch(`${API_BASE}/workouts`, {
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    state.workoutRoutines = Array.isArray(data) ? data : [];
    renderRoutines();
  } catch (err) {
    console.error('Error fetching workout routines:', err);
    state.workoutRoutines = [];
  }
}

export async function fetchPlans() {
  try {
    const res = await fetch(`${API_BASE}/plans`, {
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    state.workoutPlans = Array.isArray(data) ? data : [];
    renderPlans();
  } catch (err) {
    console.error('Error fetching plans:', err);
    state.workoutPlans = [];
  }
}

export async function fetchHistory() {
  try {
    const res = await fetch(`${API_BASE}/history`, {
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    state.historyLogs = Array.isArray(data) ? data : [];
    renderHistory();
  } catch (err) {
    console.error('Error fetching history:', err);
    state.historyLogs = [];
  }
}

export async function copyWorkoutTemplate(workoutId) {
  try {
    const res = await fetch(`${API_BASE}/workouts/copy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify({ workoutId })
    });
    if (!res.ok) throw new Error('Failed to copy workout template');
    await loadDashboardData();
  } catch (err) {
    alert(err.message);
  }
}

export async function copyPlan(planId) {
  try {
    const res = await fetch(`${API_BASE}/plans/copy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify({ planId })
    });
    if (!res.ok) throw new Error('Failed to copy training plan');
    await loadDashboardData();
  } catch (err) {
    alert(err.message);
  }
}

export async function toggleSoftDelete(id) {
  try {
    await fetch(`${API_BASE}/exercises?id=${id}`, {
      method: 'DELETE',
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    await fetchExercises();
  } catch (err) {
    console.error('Error soft deleting exercise:', err);
  }
}

export async function deleteRoutine(id) {
  try {
    await fetch(`${API_BASE}/workouts?id=${id}`, {
      method: 'DELETE',
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    await fetchWorkouts();
  } catch (err) {
    console.error('Error deleting routine:', err);
  }
}

export async function toggleWorkoutExecuted(planId, workoutId, isExecuted) {
  try {
    await fetch(`${API_BASE}/plans`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': state.activeSessionId
      },
      body: JSON.stringify({ planId, workoutId, isExecuted })
    });
    await fetchPlans();
  } catch (err) {
    console.error('Error toggling workout executed:', err);
  }
}

export async function exportVault() {
  try {
    const res = await fetch(`${API_BASE}/vault/export`, {
      headers: { 'X-Session-ID': state.activeSessionId }
    });
    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `gymrat_vault_${state.activeSessionId.substring(0, 8)}.json`;
    document.body.appendChild(a);
    a.click();
    a.remove();
  } catch (err) {
    console.error('Error exporting vault:', err);
  }
}
