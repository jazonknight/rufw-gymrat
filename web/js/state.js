/**
 * GymRat SPA State Management Module
 * Holds active workspace session ID, in-memory catalogs, live logger state, and application constants.
 */

export const API_BASE = '/api';
export const APP_VERSION = 'v2.0.5';

export const state = {
  activeSessionId: localStorage.getItem('gymrat_session_id') || '',
  exerciseCatalog: [],
  workoutRoutines: [],
  workoutPlans: [],
  historyLogs: [],
  
  // Live Workout Session Logger State
  liveWorkoutSession: null,
  liveTimerInterval: null,
  liveTimerSeconds: 0,
  restTimerInterval: null
};

/**
 * Escapes HTML characters to prevent XSS in rendered card titles and descriptions.
 * @param {string} str 
 * @returns {string}
 */
export function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/[&<>"']/g, function (m) {
    return {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#039;'
    }[m];
  });
}

/**
 * Updates the header badge and session info UI elements.
 * @param {string} id 
 */
export function updateSessionUI(id) {
  const el = document.getElementById('active-session-id');
  if (el) {
    el.textContent = id ? id.substring(0, 8) + '...' : 'Initializing...';
  }
}
