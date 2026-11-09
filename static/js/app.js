/**
 * TaskFlow — Modern Go Task Management Frontend
 * Clean, vanilla JavaScript managing state, UI interactions, and REST API communication.
 */

(() => {
  'use strict';

  // =========================================================================
  // Application State
  // =========================================================================
  const state = {
    tasks: [],
    filter: 'all',          // 'all' | 'active' | 'completed'
    searchQuery: '',
    sortBy: 'newest',       // 'newest' | 'oldest' | 'priority' | 'due_date'
    isLoading: false,
    editingTaskId: null,
  };

  // =========================================================================
  // DOM Elements
  // =========================================================================
  const elements = {
    // Header & Status
    themeToggleBtn: document.getElementById('theme-toggle-btn'),
    apiStatusPill: document.getElementById('api-status-pill'),
    apiStatusText: document.getElementById('api-status-text'),

    // Stats
    statTotal: document.getElementById('stat-total'),
    statCompleted: document.getElementById('stat-completed'),
    statPending: document.getElementById('stat-pending'),
    statRate: document.getElementById('stat-rate'),
    progressBarFill: document.getElementById('progress-bar-fill'),

    // Task Creation Form
    createForm: document.getElementById('create-task-form'),
    titleInput: document.getElementById('task-title-input'),
    descInput: document.getElementById('task-desc-input'),
    prioritySelect: document.getElementById('task-priority-select'),
    dueInput: document.getElementById('task-due-input'),
    addTaskBtn: document.getElementById('add-task-btn'),

    // Toolbar & Controls
    filterTabs: document.querySelectorAll('.filter-tab'),
    badgeAll: document.getElementById('badge-all'),
    badgeActive: document.getElementById('badge-active'),
    badgeCompleted: document.getElementById('badge-completed'),
    searchInput: document.getElementById('search-input'),
    clearSearchBtn: document.getElementById('clear-search-btn'),
    sortSelect: document.getElementById('sort-select'),
    clearCompletedBtn: document.getElementById('clear-completed-btn'),

    // List & States
    taskList: document.getElementById('task-list'),
    loadingState: document.getElementById('loading-state'),
    emptyState: document.getElementById('empty-state'),
    emptyTitle: document.getElementById('empty-title'),
    emptyDesc: document.getElementById('empty-desc'),

    // Edit Modal
    editBackdrop: document.getElementById('edit-modal-backdrop'),
    editForm: document.getElementById('edit-task-form'),
    editTaskId: document.getElementById('edit-task-id'),
    editTitleInput: document.getElementById('edit-title-input'),
    editDescInput: document.getElementById('edit-desc-input'),
    editPrioritySelect: document.getElementById('edit-priority-select'),
    editDueInput: document.getElementById('edit-due-input'),
    modalCloseBtn: document.getElementById('modal-close-btn'),
    editCancelBtn: document.getElementById('edit-cancel-btn'),

    // Toast Container
    toastContainer: document.getElementById('toast-container'),
  };

  // =========================================================================
  // API Service
  // =========================================================================
  const API_BASE = '/api';

  async function apiRequest(endpoint, options = {}) {
    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...(options.headers || {}),
        },
        ...options,
      });

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        throw new Error(data.error || `Server responded with ${response.status}`);
      }

      setBackendOnline(true);
      return data;
    } catch (err) {
      if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
        setBackendOnline(false);
      }
      throw err;
    }
  }

  // =========================================================================
  // Backend Connection Health Indicator
  // =========================================================================
  function setBackendOnline(isOnline) {
    if (isOnline) {
      elements.apiStatusPill.classList.remove('offline');
      elements.apiStatusText.textContent = 'API Connected';
    } else {
      elements.apiStatusPill.classList.add('offline');
      elements.apiStatusText.textContent = 'Disconnected';
    }
  }

  async function checkServerHealth() {
    try {
      await apiRequest('/health');
      setBackendOnline(true);
    } catch {
      setBackendOnline(false);
    }
  }

  // =========================================================================
  // Data Fetching & Sync
  // =========================================================================
  async function loadTasksAndStats() {
    state.isLoading = true;
    renderUI();

    try {
      // Build query string based on active filters and sort
      const params = new URLSearchParams({
        status: state.filter,
        sort: state.sortBy,
      });

      if (state.searchQuery.trim()) {
        params.append('search', state.searchQuery.trim());
      }

      // Fetch tasks and stats in parallel
      const [tasksRes, statsRes] = await Promise.all([
        apiRequest(`/tasks?${params.toString()}`),
        apiRequest('/tasks/stats'),
      ]);

      state.tasks = tasksRes.data || [];
      updateStatsUI(statsRes.data || { total: 0, completed: 0, pending: 0, completion_rate: 0 });
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      state.isLoading = false;
      renderUI();
    }
  }

  async function refreshStatsOnly() {
    try {
      const statsRes = await apiRequest('/tasks/stats');
      updateStatsUI(statsRes.data || { total: 0, completed: 0, pending: 0, completion_rate: 0 });
    } catch (err) {
      console.warn('Failed to update stats:', err);
    }
  }

  // =========================================================================
  // Task Operations (CRUD)
  // =========================================================================

  // 1. Create Task
  async function handleCreateTask(e) {
    e.preventDefault();

    const title = elements.titleInput.value.trim();
    if (!title) {
      showToast('Please enter a task title', 'error');
      elements.titleInput.focus();
      return;
    }

    const payload = {
      title: title,
      description: elements.descInput.value.trim(),
      priority: elements.prioritySelect.value,
      due_date: elements.dueInput.value,
    };

    elements.addTaskBtn.disabled = true;

    try {
      const res = await apiRequest('/tasks', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      showToast('Task added successfully! 🎉', 'success');

      // Reset form
      elements.createForm.reset();
      elements.prioritySelect.value = 'medium';

      // Reload list & stats
      await loadTasksAndStats();
      elements.titleInput.focus();
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      elements.addTaskBtn.disabled = false;
    }
  }

  // 2. Toggle Task Complete
  async function handleToggleTask(id) {
    try {
      const res = await apiRequest(`/tasks/${id}/toggle`, {
        method: 'PATCH',
      });

      const updatedTask = res.data;
      const isDone = updatedTask.completed;

      showToast(isDone ? 'Task completed! ✨' : 'Task marked active', 'info');

      // If we are currently filtering by active or completed, reload tasks
      if (state.filter !== 'all') {
        await loadTasksAndStats();
      } else {
        // In-place state update for snappy UI
        const index = state.tasks.findIndex((t) => t.id === id);
        if (index !== -1) {
          state.tasks[index] = updatedTask;
          renderTaskList();
        }
        await refreshStatsOnly();
      }
    } catch (err) {
      showToast(err.message, 'error');
    }
  }

  // 3. Delete Task
  async function handleDeleteTask(id, title) {
    const confirmed = confirm(`Are you sure you want to delete "${title}"?`);
    if (!confirmed) return;

    try {
      await apiRequest(`/tasks/${id}`, {
        method: 'DELETE',
      });

      showToast('Task deleted', 'info');
      await loadTasksAndStats();
    } catch (err) {
      showToast(err.message, 'error');
    }
  }

  // 4. Update Task (via Edit Modal)
  async function handleSaveEdit(e) {
    e.preventDefault();

    const id = elements.editTaskId.value;
    const title = elements.editTitleInput.value.trim();

    if (!title) {
      showToast('Task title cannot be empty', 'error');
      elements.editTitleInput.focus();
      return;
    }

    const payload = {
      title: title,
      description: elements.editDescInput.value.trim(),
      priority: elements.editPrioritySelect.value,
      due_date: elements.editDueInput.value,
    };

    try {
      await apiRequest(`/tasks/${id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });

      showToast('Task updated successfully!', 'success');
      closeEditModal();
      await loadTasksAndStats();
    } catch (err) {
      showToast(err.message, 'error');
    }
  }

  // 5. Clear Completed Tasks
  async function handleClearCompleted() {
    const completedCount = parseInt(elements.statCompleted.textContent, 10) || 0;
    if (completedCount === 0) {
      showToast('No completed tasks to clear', 'info');
      return;
    }

    const confirmed = confirm(`Delete all ${completedCount} completed task(s)? This cannot be undone.`);
    if (!confirmed) return;

    try {
      const res = await apiRequest('/tasks/completed', {
        method: 'DELETE',
      });

      const deleted = res.data?.deleted_count || completedCount;
      showToast(`Cleared ${deleted} completed task(s)`, 'success');
      await loadTasksAndStats();
    } catch (err) {
      showToast(err.message, 'error');
    }
  }

  // =========================================================================
  // Edit Modal Handling
  // =========================================================================
  function openEditModal(task) {
    state.editingTaskId = task.id;
    elements.editTaskId.value = task.id;
    elements.editTitleInput.value = task.title || '';
    elements.editDescInput.value = task.description || '';
    elements.editPrioritySelect.value = task.priority || 'medium';
    elements.editDueInput.value = task.due_date || '';

    elements.editBackdrop.style.display = 'flex';
    elements.editBackdrop.setAttribute('aria-hidden', 'false');
    elements.editTitleInput.focus();
  }

  function closeEditModal() {
    elements.editBackdrop.style.display = 'none';
    elements.editBackdrop.setAttribute('aria-hidden', 'true');
    elements.editForm.reset();
    state.editingTaskId = null;
  }

  // =========================================================================
  // UI Rendering & Helpers
  // =========================================================================
  function escapeHTML(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function formatRelativeTime(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '';

    const now = new Date();
    const diffSeconds = Math.floor((now - date) / 1000);

    if (diffSeconds < 60) return 'Just now';
    if (diffSeconds < 3600) {
      const mins = Math.floor(diffSeconds / 60);
      return `${mins}m ago`;
    }
    if (diffSeconds < 86400) {
      const hours = Math.floor(diffSeconds / 3600);
      return `${hours}h ago`;
    }
    const days = Math.floor(diffSeconds / 86400);
    if (days === 1) return 'Yesterday';
    if (days < 7) return `${days}d ago`;

    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  }

  function formatDueDate(dueDateStr) {
    if (!dueDateStr) return null;
    const parts = dueDateStr.split('-');
    if (parts.length !== 3) return { text: dueDateStr, status: 'normal' };

    const due = new Date(parts[0], parts[1] - 1, parts[2]);
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const diffDays = Math.round((due - today) / (1000 * 60 * 60 * 24));

    if (diffDays < 0) {
      return { text: `Overdue (${Math.abs(diffDays)}d)`, status: 'overdue' };
    }
    if (diffDays === 0) {
      return { text: 'Due Today', status: 'due-today' };
    }
    if (diffDays === 1) {
      return { text: 'Due Tomorrow', status: 'normal' };
    }
    return {
      text: due.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      status: 'normal',
    };
  }

  function renderTaskList() {
    elements.taskList.innerHTML = '';

    if (state.tasks.length === 0) {
      elements.emptyState.style.display = 'flex';

      if (state.searchQuery.trim()) {
        elements.emptyTitle.textContent = 'No matching tasks';
        elements.emptyDesc.textContent = `No tasks found matching "${state.searchQuery}". Try a different search term.`;
      } else if (state.filter === 'active') {
        elements.emptyTitle.textContent = 'No active tasks';
        elements.emptyDesc.textContent = 'All tasks are completed! Enjoy your day or add new tasks above.';
      } else if (state.filter === 'completed') {
        elements.emptyTitle.textContent = 'No completed tasks yet';
        elements.emptyDesc.textContent = 'Mark tasks as completed when you finish them.';
      } else {
        elements.emptyTitle.textContent = 'Your task list is empty';
        elements.emptyDesc.textContent = 'Create your first task using the form above to get started!';
      }
      return;
    }

    elements.emptyState.style.display = 'none';

    state.tasks.forEach((task) => {
      const li = document.createElement('li');
      li.className = `task-item task-item-enter ${task.completed ? 'completed' : ''}`;
      li.dataset.id = task.id;

      // Checkbox
      const checkBtn = document.createElement('button');
      checkBtn.className = 'task-checkbox-btn';
      checkBtn.setAttribute('aria-label', task.completed ? 'Mark task as incomplete' : 'Mark task as complete');
      checkBtn.title = task.completed ? 'Mark active' : 'Mark completed';
      checkBtn.innerHTML = `
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      `;
      checkBtn.addEventListener('click', () => handleToggleTask(task.id));

      // Content container
      const contentDiv = document.createElement('div');
      contentDiv.className = 'task-content';

      // Title
      const titleEl = document.createElement('h3');
      titleEl.className = 'task-title';
      titleEl.textContent = task.title;

      contentDiv.appendChild(titleEl);

      // Description (if present)
      if (task.description) {
        const descEl = document.createElement('p');
        descEl.className = 'task-description';
        descEl.textContent = task.description;
        contentDiv.appendChild(descEl);
      }

      // Meta row (Priority, Due Date, Created Time)
      const metaDiv = document.createElement('div');
      metaDiv.className = 'task-meta';

      // Priority badge
      const priority = task.priority || 'medium';
      const priorityBadge = document.createElement('span');
      priorityBadge.className = `badge badge-priority-${priority}`;
      priorityBadge.innerHTML = `
        <svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2.5">
          <circle cx="12" cy="12" r="8"></circle>
        </svg>
        ${priority.charAt(0).toUpperCase() + priority.slice(1)}
      `;
      metaDiv.appendChild(priorityBadge);

      // Due date badge (if present)
      const dueInfo = formatDueDate(task.due_date);
      if (dueInfo) {
        const dueBadge = document.createElement('span');
        dueBadge.className = `badge badge-due ${dueInfo.status}`;
        dueBadge.innerHTML = `
          <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
          </svg>
          ${dueInfo.text}
        `;
        metaDiv.appendChild(dueBadge);
      }

      // Created time
      const timeStr = formatRelativeTime(task.created_at);
      if (timeStr) {
        const timeEl = document.createElement('span');
        timeEl.className = 'task-time';
        timeEl.textContent = `• Created ${timeStr}`;
        metaDiv.appendChild(timeEl);
      }

      contentDiv.appendChild(metaDiv);

      // Action buttons (Edit & Delete)
      const actionsDiv = document.createElement('div');
      actionsDiv.className = 'task-actions';

      // Edit button
      const editBtn = document.createElement('button');
      editBtn.className = 'action-btn action-btn-edit';
      editBtn.setAttribute('aria-label', `Edit task: ${task.title}`);
      editBtn.title = 'Edit task';
      editBtn.innerHTML = `
        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 20h9"></path>
          <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
        </svg>
      `;
      editBtn.addEventListener('click', () => openEditModal(task));

      // Delete button
      const delBtn = document.createElement('button');
      delBtn.className = 'action-btn action-btn-delete';
      delBtn.setAttribute('aria-label', `Delete task: ${task.title}`);
      delBtn.title = 'Delete task';
      delBtn.innerHTML = `
        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
      `;
      delBtn.addEventListener('click', () => handleDeleteTask(task.id, task.title));

      actionsDiv.appendChild(editBtn);
      actionsDiv.appendChild(delBtn);

      // Assemble list item
      li.appendChild(checkBtn);
      li.appendChild(contentDiv);
      li.appendChild(actionsDiv);

      elements.taskList.appendChild(li);
    });
  }

  function updateStatsUI(stats) {
    const total = stats.total || 0;
    const completed = stats.completed || 0;
    const pending = stats.pending || 0;
    const rate = stats.completion_rate || 0;

    elements.statTotal.textContent = total;
    elements.statCompleted.textContent = completed;
    elements.statPending.textContent = pending;
    elements.statRate.textContent = `${rate}%`;

    elements.progressBarFill.style.width = `${rate}%`;
    const wrapper = document.getElementById('progress-bar-wrapper');
    if (wrapper) {
      wrapper.setAttribute('aria-valuenow', rate);
    }

    elements.badgeAll.textContent = total;
    elements.badgeActive.textContent = pending;
    elements.badgeCompleted.textContent = completed;
  }

  function renderUI() {
    if (state.isLoading) {
      elements.loadingState.style.display = 'flex';
      elements.emptyState.style.display = 'none';
      elements.taskList.innerHTML = '';
    } else {
      elements.loadingState.style.display = 'none';
      renderTaskList();
    }
  }

  // =========================================================================
  // Toast Notification System
  // =========================================================================
  function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    let iconSvg = '';
    if (type === 'success') {
      iconSvg = '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>';
    } else if (type === 'error') {
      iconSvg = '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>';
    } else {
      iconSvg = '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>';
    }

    toast.innerHTML = `
      ${iconSvg}
      <span>${escapeHTML(message)}</span>
    `;

    toast.addEventListener('click', () => {
      toast.classList.add('toast-closing');
      setTimeout(() => toast.remove(), 250);
    });

    elements.toastContainer.appendChild(toast);

    setTimeout(() => {
      if (toast.parentElement) {
        toast.classList.add('toast-closing');
        setTimeout(() => toast.remove(), 250);
      }
    }, 3200);
  }

  // =========================================================================
  // Theme Toggle (Dark / Light)
  // =========================================================================
  function initTheme() {
    const savedTheme = localStorage.getItem('taskflow_theme');
    const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light');

    applyTheme(initialTheme);
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('taskflow_theme', theme);
  }

  function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    const nextTheme = currentTheme === 'light' ? 'dark' : 'light';
    applyTheme(nextTheme);
  }

  // =========================================================================
  // Event Listeners Setup
  // =========================================================================
  function initEventListeners() {
    // Theme toggle
    elements.themeToggleBtn.addEventListener('click', toggleTheme);

    // Create form
    elements.createForm.addEventListener('submit', handleCreateTask);

    // Filter tabs
    elements.filterTabs.forEach((tab) => {
      tab.addEventListener('click', () => {
        elements.filterTabs.forEach((t) => {
          t.classList.remove('active');
          t.setAttribute('aria-selected', 'false');
        });
        tab.classList.add('active');
        tab.setAttribute('aria-selected', 'true');

        state.filter = tab.dataset.status;
        loadTasksAndStats();
      });
    });

    // Search input (debounce 250ms)
    let searchTimeout = null;
    elements.searchInput.addEventListener('input', (e) => {
      const val = e.target.value;
      elements.clearSearchBtn.style.display = val ? 'block' : 'none';

      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => {
        state.searchQuery = val;
        loadTasksAndStats();
      }, 250);
    });

    // Clear search
    elements.clearSearchBtn.addEventListener('click', () => {
      elements.searchInput.value = '';
      elements.clearSearchBtn.style.display = 'none';
      state.searchQuery = '';
      loadTasksAndStats();
      elements.searchInput.focus();
    });

    // Sort select
    elements.sortSelect.addEventListener('change', (e) => {
      state.sortBy = e.target.value;
      loadTasksAndStats();
    });

    // Clear completed button
    elements.clearCompletedBtn.addEventListener('click', handleClearCompleted);

    // Edit modal form submit
    elements.editForm.addEventListener('submit', handleSaveEdit);
    elements.modalCloseBtn.addEventListener('click', closeEditModal);
    elements.editCancelBtn.addEventListener('click', closeEditModal);

    // Close modal on backdrop click
    elements.editBackdrop.addEventListener('click', (e) => {
      if (e.target === elements.editBackdrop) {
        closeEditModal();
      }
    });

    // Global keyboard shortcuts (Escape closes modal)
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && elements.editBackdrop.style.display === 'flex') {
        closeEditModal();
      }
    });

    // Periodically verify server health every 30s
    setInterval(checkServerHealth, 30000);
  }

  // =========================================================================
  // App Initialization
  // =========================================================================
  function init() {
    initTheme();
    initEventListeners();
    loadTasksAndStats();
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();

