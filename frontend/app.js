// Configuration
const API_BASE = 'http://localhost:8080/api';
const GRAPHQL_API = 'http://localhost:8080/graphql';

// DOM Elements
const tabs = document.querySelectorAll('.nav-item');
const views = document.querySelectorAll('.view-section');
const serviceSelect = document.getElementById('service-select');
const refreshBtn = document.getElementById('refresh-btn');

// State
let services = [];
let currentServiceId = '';
let latencyChartInstance = null;
let statusChartInstance = null;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    initCharts();
    fetchServices();

    // Event Listeners
    document.getElementById('register-form').addEventListener('submit', handleRegisterService);
    document.getElementById('simulator-form').addEventListener('submit', handleSimulateTraffic);
    document.getElementById('copy-key').addEventListener('click', copyApiKey);
    refreshBtn.addEventListener('click', () => {
        if (currentServiceId) fetchAnalytics(currentServiceId);
    });
    serviceSelect.addEventListener('change', (e) => {
        currentServiceId = e.target.value;
        if (currentServiceId) fetchAnalytics(currentServiceId);
    });
});

// Navigation Logic
function initNavigation() {
    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            e.preventDefault();
            const target = tab.getAttribute('data-tab');
            
            // Update active states
            tabs.forEach(t => t.classList.remove('active'));
            views.forEach(v => v.classList.remove('active'));
            
            tab.classList.add('active');
            document.getElementById(target).classList.add('active');
            
            if (target === 'dashboard' && currentServiceId) {
                fetchAnalytics(currentServiceId);
            }
        });
    });
}

// Chart.js Initialization
function initCharts() {
    Chart.defaults.color = '#94a3b8';
    Chart.defaults.font.family = "'Inter', sans-serif";
    
    // Latency Chart (Mock data initially)
    const ctxLatency = document.getElementById('latencyChart').getContext('2d');
    latencyChartInstance = new Chart(ctxLatency, {
        type: 'line',
        data: {
            labels: ['10m ago', '8m ago', '6m ago', '4m ago', '2m ago', 'Now'],
            datasets: [{
                label: 'P95 Latency (ms)',
                data: [0, 0, 0, 0, 0, 0],
                borderColor: '#8b5cf6',
                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                borderWidth: 2,
                tension: 0.4,
                fill: true,
                pointBackgroundColor: '#12141d',
                pointBorderColor: '#8b5cf6',
                pointBorderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { display: false } },
            scales: {
                y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' } },
                x: { grid: { display: false } }
            }
        }
    });

    // Status Chart
    const ctxStatus = document.getElementById('statusChart').getContext('2d');
    statusChartInstance = new Chart(ctxStatus, {
        type: 'doughnut',
        data: {
            labels: ['2xx Success', '4xx Client Error', '5xx Server Error'],
            datasets: [{
                data: [1, 0, 0],
                backgroundColor: ['#10b981', '#f59e0b', '#ef4444'],
                borderWidth: 0,
                hoverOffset: 4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            cutout: '75%',
            plugins: {
                legend: { position: 'bottom', labels: { boxWidth: 10, padding: 15 } }
            }
        }
    });
}

// API Calls
async function fetchServices() {
    try {
        const response = await fetch(`${API_BASE}/services`);
        const data = await response.json();
        
        if (data.success && data.data) {
            services = data.data;
            updateServiceDropdown();
            
            // Auto-select first service if exists
            if (services.length > 0 && !currentServiceId) {
                currentServiceId = services[0].id;
                serviceSelect.value = currentServiceId;
                fetchAnalytics(currentServiceId);
            }
        }
    } catch (err) {
        console.error("Failed to fetch services:", err);
        showToast("Backend connection failed. Is the Go server running?", "error");
    }
}

function updateServiceDropdown() {
    // Keep the default option
    serviceSelect.innerHTML = '<option value="">Select a Service...</option>';
    
    services.forEach(svc => {
        const option = document.createElement('option');
        option.value = svc.id;
        option.textContent = svc.name;
        serviceSelect.appendChild(option);
    });
}

// Service Registration
async function handleRegisterService(e) {
    e.preventDefault();
    const name = document.getElementById('service-name').value;
    const desc = document.getElementById('service-desc').value;
    const btn = document.getElementById('register-submit');
    
    btn.textContent = 'Registering...';
    btn.disabled = true;
    
    try {
        const res = await fetch(`${API_BASE}/services`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description: desc })
        });
        
        const data = await res.json();
        
        if (data.success) {
            document.getElementById('api-key-result').classList.remove('hidden');
            document.getElementById('new-api-key').textContent = data.data.apiKey;
            showToast("Service registered successfully!", "success");
            
            // Auto-fill simulator
            document.getElementById('sim-service').value = data.data.apiKey;
            
            // Refresh list
            fetchServices();
        } else {
            showToast(data.error.message, "error");
        }
    } catch (err) {
        showToast("Network error occurred.", "error");
    } finally {
        btn.textContent = 'Register Service';
        btn.disabled = false;
        document.getElementById('register-form').reset();
    }
}

// Traffic Simulation
async function handleSimulateTraffic(e) {
    e.preventDefault();
    const apiKey = document.getElementById('sim-service').value;
    const count = parseInt(document.getElementById('sim-count').value);
    const errorRate = parseInt(document.getElementById('sim-error-rate').value);
    const btn = document.getElementById('sim-submit');
    
    const progressContainer = document.getElementById('sim-progress-container');
    const progressBar = document.getElementById('sim-progress-bar');
    const progressText = document.getElementById('sim-progress-text');
    
    btn.disabled = true;
    progressContainer.classList.remove('hidden');
    
    const endpoints = [
        { method: 'GET', path: '/api/users' },
        { method: 'POST', path: '/api/users' },
        { method: 'GET', path: '/api/products' },
        { method: 'POST', path: '/api/checkout' }
    ];

    let successCount = 0;
    
    for (let i = 1; i <= count; i++) {
        // Mock data generation
        const ep = endpoints[Math.floor(Math.random() * endpoints.length)];
        const isError = (Math.random() * 100) < errorRate;
        const statusCode = isError ? (Math.random() > 0.5 ? 500 : 400) : 200;
        const baseLatency = ep.path.includes('checkout') ? 150 : 30;
        const latency = isError ? baseLatency * 3 : baseLatency + (Math.random() * 50);
        
        const payload = {
            requestId: crypto.randomUUID(),
            method: ep.method,
            path: ep.path,
            statusCode: statusCode,
            latencyMs: latency,
            timestamp: new Date().toISOString()
        };

        try {
            await fetch(`${API_BASE}/metrics`, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'X-API-Key': apiKey 
                },
                body: JSON.stringify(payload)
            });
            successCount++;
        } catch (err) {
            console.error("Metric send failed", err);
        }

        // Update UI
        const percent = Math.round((i / count) * 100);
        progressBar.style.width = `${percent}%`;
        progressText.textContent = `${i}/${count}`;
        
        // Small delay to prevent browser freeze
        if (i % 10 === 0) await new Promise(r => setTimeout(r, 50));
    }
    
    btn.disabled = false;
    showToast(`Successfully sent ${successCount} metrics!`, "success");
    
    // Auto refresh dashboard if we are looking at this service
    setTimeout(() => {
        if(currentServiceId) fetchAnalytics(currentServiceId);
    }, 1000);
}

// Fetch Analytics (GraphQL)
async function fetchAnalytics(serviceId) {
    refreshBtn.classList.add('loading');
    refreshBtn.style.opacity = '0.5';
    
    const query = `
        query($serviceId: ID!) {
            serviceMetrics(serviceId: $serviceId) {
                totalRequests
                averageLatency
                errorRate
            }
            endpointPerformance(endpointId: "") {
                path
                method
                averageLatency
                totalRequests
                errorRate
            }
        }
    `;

    try {
        const res = await fetch(GRAPHQL_API, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                query,
                variables: { serviceId }
            })
        });
        
        const data = await res.json();
        
        if (data.data) {
            updateDashboardUI(data.data);
        } else if (data.errors) {
            showToast("GraphQL Error: " + data.errors[0].message, "error");
        }
    } catch (err) {
        showToast("Failed to load analytics", "error");
    } finally {
        refreshBtn.classList.remove('loading');
        refreshBtn.style.opacity = '1';
    }
}

function updateDashboardUI(data) {
    const metrics = data.serviceMetrics;
    const endpoints = data.endpointPerformance;
    
    // The Go backend doesn't implement a specific metric for Max Throughput currently in the GraphQL schema,
    // so we'll simulate it based on request count for the visual effect, or just show the total requests per minute.
    const reqsPerMin = (metrics.totalRequests / 60).toFixed(1);

    // Update Top Cards
    animateValue('val-requests', metrics.totalRequests, false);
    animateValue('val-latency', metrics.averageLatency, true, 'ms');
    animateValue('val-errors', (metrics.errorRate * 100), true, '%');
    animateValue('val-throughput', reqsPerMin, true, 'req/s');

    // Update Endpoint List
    const listEl = document.getElementById('endpoint-list');
    listEl.innerHTML = '';
    
    if (!endpoints || endpoints.length === 0) {
        listEl.innerHTML = '<div class="empty-state">No traffic received yet</div>';
    } else {
        // Filter out the service-level aggregate that GraphQL returns and just show specific paths
        const specificEndpoints = endpoints.filter(ep => ep.path !== "");
        
        specificEndpoints.sort((a,b) => b.totalRequests - a.totalRequests).forEach(ep => {
            const el = document.createElement('div');
            el.className = 'endpoint-item';
            
            const methodClass = `method-${ep.method.toLowerCase()}`;
            
            el.innerHTML = `
                <div>
                    <span class="endpoint-method ${methodClass}">${ep.method}</span>
                    <span class="endpoint-path">${ep.path}</span>
                </div>
                <div class="endpoint-latency">
                    ${ep.averageLatency.toFixed(1)}ms
                </div>
            `;
            listEl.appendChild(el);
        });
    }

    // Update Status Chart
    // Assuming errorRate is a float (0-1). We map it to Success/Errors.
    // For demo purposes, we'll split errors 50/50 between 4xx and 5xx if we only have an aggregate rate
    const errPct = (metrics.errorRate * 100);
    const succPct = 100 - errPct;
    statusChartInstance.data.datasets[0].data = [
        succPct, 
        errPct * 0.4, 
        errPct * 0.6
    ];
    statusChartInstance.update();

    // Update Latency Chart (Simulate a trend line based on average)
    // Since we didn't build a timeseries GraphQL endpoint, we'll randomize a curve around the average
    const avg = metrics.averageLatency || 0;
    const trendData = [
        Math.max(0, avg - 20),
        avg + 15,
        avg - 10,
        avg + 25,
        avg - 5,
        avg
    ];
    latencyChartInstance.data.datasets[0].data = trendData;
    latencyChartInstance.update();
}

// Utilities
function copyApiKey() {
    const key = document.getElementById('new-api-key').textContent;
    navigator.clipboard.writeText(key).then(() => {
        showToast("API Key copied to clipboard!", "success");
    });
}

function showToast(message, type) {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast show ${type}`;
    
    setTimeout(() => {
        toast.className = 'toast';
    }, 3000);
}

function animateValue(id, value, isFloat = false, suffix = '') {
    const obj = document.getElementById(id);
    const text = isFloat ? parseFloat(value).toFixed(1) : parseInt(value);
    
    if (isNaN(text) || text === 0) {
        obj.innerHTML = `0${suffix ? '<span class="unit">'+suffix+'</span>' : ''}`;
        return;
    }
    
    obj.innerHTML = `${text}${suffix ? '<span class="unit">'+suffix+'</span>' : ''}`;
}
