const API_BASE = '/api';
let authToken = null;

const api = {
    async request(url, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (authToken) {
            headers['Authorization'] = `Token ${authToken}`;
        }

        try {
            const response = await fetch(`${API_BASE}${url}`, {
                ...options,
                headers
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP ${response.status}`);
            }

            const contentType = response.headers.get('content-type');
            const text = await response.text();
            
            if (contentType && contentType.includes('application/json')) {
                try {
                    const data = JSON.parse(text);
                    console.log('API Response:', url, data);
                    return data;
                } catch (e) {
                    console.log('Failed to parse JSON, returning text');
                    return text;
                }
            }
            
            if (text.trim().startsWith('{') || text.trim().startsWith('[')) {
                try {
                    const data = JSON.parse(text);
                    console.log('API Response (parsed as JSON):', url, data);
                    return data;
                } catch (e) {
                    console.log('API Response (text):', url, text);
                    return text;
                }
            }
            
            console.log('API Response (text):', url, text);
            return text;
        } catch (error) {
            throw error;
        }
    },

    async login(username, password) {
        const token = await this.request('/login', {
            method: 'POST',
            body: JSON.stringify({ name: username, password })
        });
        return token;
    },

    async ping() {
        return await this.request('/ping');
    },

    async search(phrase, limit, useIndex = false) {
        const endpoint = useIndex ? '/isearch' : '/search';
        return await this.request(`${endpoint}?phrase=${encodeURIComponent(phrase)}&limit=${limit}`);
    },

    async getStats() {
        return await this.request('/db/stats');
    },

    async getStatus() {
        return await this.request('/db/status');
    },

    async updateDB() {
        return await this.request('/db/update', { method: 'POST' });
    },

    async dropDB() {
        return await this.request('/db', { method: 'DELETE' });
    }
};

document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('login-error');

    try {
        authToken = await api.login(username, password);
        errorDiv.style.display = 'none';
        document.getElementById('login-section').style.display = 'none';
        document.getElementById('main-section').style.display = 'block';
        checkServicesStatus();
    } catch (error) {
        errorDiv.textContent = `Ошибка авторизации: ${error.message}`;
        errorDiv.style.display = 'block';
    }
});

document.getElementById('check-status-btn').addEventListener('click', checkServicesStatus);

async function checkServicesStatus() {
    try {
        const status = await api.ping();
        console.log('Ping status:', status);
        const statusDiv = document.getElementById('services-status');
        statusDiv.innerHTML = '';

        const replies = status.replies || status;
        if (replies && typeof replies === 'object') {
            for (const [service, state] of Object.entries(replies)) {
                const div = document.createElement('div');
                div.className = `service-status ${state === 'ok' ? 'ok' : 'unavailable'}`;
                div.textContent = `${service}: ${state === 'ok' ? '✓ OK' : '✗ Недоступен'}`;
                statusDiv.appendChild(div);
            }
        }
    } catch (error) {
        console.error('Failed to check services status:', error);
    }
}

document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        const tab = btn.dataset.tab;
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        btn.classList.add('active');
        document.getElementById(`${tab}-tab`).classList.add('active');
        
        if (tab === 'stats') {
            setTimeout(() => {
                loadStats();
                loadUpdateStatus();
            }, 100);
        }
    });
});

document.getElementById('search-btn').addEventListener('click', async () => {
    const phrase = document.getElementById('search-phrase').value;
    const limit = parseInt(document.getElementById('search-limit').value) || 10;
    const useIndex = document.querySelector('input[name="search-type"]:checked').value === 'index';
    const resultsDiv = document.getElementById('search-results');

    if (!phrase.trim()) {
        resultsDiv.innerHTML = '<div class="error">Введите фразу для поиска</div>';
        return;
    }

    resultsDiv.innerHTML = '<div class="loading">Поиск</div>';

    try {
        const result = await api.search(phrase, limit, useIndex);
        console.log('Search result:', result);
        console.log('Result type:', typeof result);
        console.log('Result.comics:', result?.comics);
        console.log('Result.comics length:', result?.comics?.length);
        console.log('resultsDiv element:', resultsDiv);
        
        if (result && result.comics && Array.isArray(result.comics) && result.comics.length > 0) {
            let html = '<h3>Найдено комиксов: ' + (result.total || result.comics.length) + '</h3>';
            
            result.comics.forEach(comic => {
                html += '<div class="comic-card">';
                html += '<h3>Комикс #' + comic.id + '</h3>';
                html += '<a href="' + comic.url + '" target="_blank">' + comic.url + '</a>';
                html += '</div>';
            });
            
            console.log('Setting HTML, length:', html.length);
            resultsDiv.innerHTML = html;
            console.log('HTML set, resultsDiv.innerHTML length:', resultsDiv.innerHTML.length);
            
            setTimeout(() => {
                console.log('After timeout, resultsDiv.innerHTML length:', resultsDiv.innerHTML.length);
                console.log('resultsDiv visible:', resultsDiv.offsetParent !== null);
            }, 100);
        } else {
            console.log('No comics found or empty result');
            resultsDiv.innerHTML = '<div class="message">Комиксы не найдены</div>';
        }
    } catch (error) {
        console.error('Search error:', error);
        resultsDiv.innerHTML = `<div class="error">Ошибка поиска: ${error.message}</div>`;
    }
});

document.getElementById('refresh-stats-btn').addEventListener('click', async () => {
    await loadStats();
    await loadUpdateStatus();
});

async function loadStats() {
    const statsDiv = document.getElementById('stats-content');
    statsDiv.innerHTML = '<div class="loading">Загрузка статистики</div>';

    try {
        const stats = await api.getStats();
        console.log('Stats received:', stats);
        console.log('Stats type:', typeof stats);
        console.log('Stats keys:', stats ? Object.keys(stats) : 'null');
        
        if (stats && typeof stats === 'object') {
            statsDiv.innerHTML = `
                <div class="stat-card">
                    <h3>Статистика базы данных</h3>
                    <div class="stat-item">
                        <span class="stat-label">Всего комиксов в XKCD:</span>
                        <span class="stat-value">${stats.comics_total || 0}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Загружено комиксов:</span>
                        <span class="stat-value">${stats.comics_fetched || 0}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Всего слов:</span>
                        <span class="stat-value">${stats.words_total || 0}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Уникальных слов:</span>
                        <span class="stat-value">${stats.words_unique || 0}</span>
                    </div>
                </div>
            `;
            console.log('Stats HTML set');
        } else {
            console.error('Invalid stats format:', stats);
            statsDiv.innerHTML = '<div class="error">Неверный формат данных статистики</div>';
        }
    } catch (error) {
        console.error('Stats error:', error);
        statsDiv.innerHTML = `<div class="error">Ошибка загрузки статистики: ${error.message}</div>`;
    }
}

async function loadUpdateStatus() {
    const statusDiv = document.getElementById('update-status-content');
    try {
        const status = await api.getStatus();
        statusDiv.innerHTML = `
            <div class="stat-card">
                <h3>Статус обновления</h3>
                <div class="stat-item">
                    <span class="stat-label">Статус:</span>
                    <span class="stat-value">${status.status === 'running' ? '🔄 Выполняется' : '✅ Простаивает'}</span>
                </div>
            </div>
        `;
    } catch (error) {
        statusDiv.innerHTML = `<div class="error">Ошибка загрузки статуса: ${error.message}</div>`;
    }
}

document.getElementById('update-db-btn').addEventListener('click', async () => {
    const messageDiv = document.getElementById('admin-message');
    messageDiv.innerHTML = '<div class="loading">Запуск обновления БД</div>';
    messageDiv.style.display = 'block';

    try {
        await api.updateDB();
        messageDiv.innerHTML = '<div class="message success">Обновление БД запущено успешно</div>';
        setTimeout(() => {
            loadStats();
            loadUpdateStatus();
        }, 2000);
    } catch (error) {
        if (error.message.includes('already runs')) {
            messageDiv.innerHTML = '<div class="message error">Обновление уже выполняется</div>';
        } else {
            messageDiv.innerHTML = `<div class="message error">Ошибка: ${error.message}</div>`;
        }
    }
});

document.getElementById('drop-db-btn').addEventListener('click', async () => {
    if (!confirm('Вы уверены, что хотите очистить базу данных? Это действие нельзя отменить!')) {
        return;
    }

    const messageDiv = document.getElementById('admin-message');
    messageDiv.innerHTML = '<div class="loading">Очистка БД</div>';
    messageDiv.style.display = 'block';

    try {
        await api.dropDB();
        messageDiv.innerHTML = '<div class="message success">База данных очищена успешно</div>';
        setTimeout(() => {
            loadStats();
            loadUpdateStatus();
        }, 1000);
    } catch (error) {
        messageDiv.innerHTML = `<div class="message error">Ошибка: ${error.message}</div>`;
    }
});

if (document.getElementById('stats-tab').classList.contains('active')) {
    loadStats();
    loadUpdateStatus();
}

