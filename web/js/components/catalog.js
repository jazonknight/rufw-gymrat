/**
 * Exercise Library Catalog UI Component
 * Handles search filtering, version history badges, soft deletion toggle, and catalog rendering.
 */

import { state, escapeHtml } from '../state.js';
import { toggleSoftDelete } from '../api.js';

export function renderCatalog() {
  const grid = document.getElementById('catalog-grid');
  if (!grid) return;
  grid.innerHTML = '';

  const showRemoved = document.getElementById('toggle-removed')?.checked ?? true;
  const searchQuery = (document.getElementById('search-exercise')?.value || '').toLowerCase().trim();
  const selectedCategory = document.getElementById('filter-category')?.value || '';

  // Get highest version of each exercise ID to handle immutable versioning
  const latestVersions = {};
  state.exerciseCatalog.forEach(ex => {
    if (!latestVersions[ex.id] || ex.version > latestVersions[ex.id].version) {
      latestVersions[ex.id] = ex;
    }
  });

  const displayList = Object.values(latestVersions).filter(ex => {
    if (!showRemoved && ex.isRemoved) return false;
    if (selectedCategory && ex.category !== selectedCategory) return false;
    if (searchQuery) {
      const matchName = (ex.name || '').toLowerCase().includes(searchQuery);
      const matchDesc = (ex.description || '').toLowerCase().includes(searchQuery);
      const matchCat = (ex.category || '').toLowerCase().includes(searchQuery);
      if (!matchName && !matchDesc && !matchCat) return false;
    }
    return true;
  });

  if (displayList.length === 0) {
    grid.innerHTML = '<p class="subtitle">No exercises matching your filter criteria.</p>';
    return;
  }

  displayList.forEach(ex => {
    const card = document.createElement('div');
    card.className = `exercise-card ${ex.isRemoved ? 'soft-deleted' : ''}`;
    
    card.innerHTML = `
      <div class="card-top">
        <h3 class="ex-title">${escapeHtml(ex.name)}</h3>
        <span class="tag-version">v${ex.version}</span>
      </div>
      <p class="ex-category">📂 ${escapeHtml(ex.category || 'General')}</p>
      <p class="ex-desc">${escapeHtml(ex.description || 'No description provided.')}</p>
      <div class="card-actions">
        <button class="btn btn-secondary btn-sm btn-bump-version" data-id="${ex.id}">Bump Version</button>
        <button class="btn btn-secondary btn-sm btn-toggle-delete" data-id="${ex.id}">${ex.isRemoved ? 'Restore' : 'Soft Delete'}</button>
      </div>
    `;

    card.querySelector('.btn-bump-version')?.addEventListener('click', () => openUpdateExerciseModal(ex.id));
    card.querySelector('.btn-toggle-delete')?.addEventListener('click', () => toggleSoftDelete(ex.id));

    grid.appendChild(card);
  });
}

export function updateCategoryDropdownOptions() {
  const catSelect = document.getElementById('filter-category');
  if (!catSelect) return;

  const currentVal = catSelect.value;
  const categories = new Set();
  state.exerciseCatalog.forEach(ex => {
    if (ex.category) categories.add(ex.category);
  });

  let html = '<option value="">All Categories</option>';
  Array.from(categories).sort().forEach(cat => {
    html += `<option value="${escapeHtml(cat)}">${escapeHtml(cat)}</option>`;
  });
  catSelect.innerHTML = html;
  catSelect.value = currentVal;
}

export function openUpdateExerciseModal(id) {
  const ex = state.exerciseCatalog.find(e => e.id === id);
  if (!ex) return;

  document.getElementById('ex-id').value = ex.id;
  document.getElementById('modal-exercise-title').textContent = `Update Exercise (Bumps to v${ex.version + 1})`;
  document.getElementById('ex-name').value = ex.name;
  document.getElementById('ex-category').value = ex.category || '';
  document.getElementById('ex-desc').value = ex.description || '';
  document.getElementById('modal-exercise')?.classList.add('active');
}

export function initFilters() {
  document.getElementById('search-exercise')?.addEventListener('input', renderCatalog);
  document.getElementById('filter-category')?.addEventListener('change', renderCatalog);
}
