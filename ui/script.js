// ========================================
// CONFIGURATION
// ========================================
const API_BASE_URL = 'http://localhost:8080/api/v1';

// ========================================
// DOM ELEMENTS
// ========================================
const elements = {
    loginForm: document.getElementById('loginForm'),
    submitBtn: document.getElementById('submitBtn'),
    identifierInput: document.getElementById('identifier'),
    passwordInput: document.getElementById('password'),
    togglePassword: document.getElementById('togglePassword'),
    alertContainer: document.getElementById('alertContainer'),
    particlesCanvas: document.getElementById('particles')
};

// ========================================
// PARTICLE ANIMATION
// ========================================
class ParticleSystem {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.particles = [];
        this.particleCount = 80;
        this.connectionDistance = 150;

        this.resize();
        this.init();
        this.animate();

        window.addEventListener('resize', () => this.resize());
    }

    resize() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }

    init() {
        this.particles = [];
        for (let i = 0; i < this.particleCount; i++) {
            this.particles.push({
                x: Math.random() * this.canvas.width,
                y: Math.random() * this.canvas.height,
                vx: (Math.random() - 0.5) * 0.5,
                vy: (Math.random() - 0.5) * 0.5,
                radius: Math.random() * 2 + 1
            });
        }
    }

    animate() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Update and draw particles
        this.particles.forEach((particle, i) => {
            // Update position
            particle.x += particle.vx;
            particle.y += particle.vy;

            // Bounce off edges
            if (particle.x < 0 || particle.x > this.canvas.width) particle.vx *= -1;
            if (particle.y < 0 || particle.y > this.canvas.height) particle.vy *= -1;

            // Draw particle
            this.ctx.beginPath();
            this.ctx.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2);
            this.ctx.fillStyle = 'rgba(102, 126, 234, 0.5)';
            this.ctx.fill();

            // Draw connections
            for (let j = i + 1; j < this.particles.length; j++) {
                const dx = this.particles[j].x - particle.x;
                const dy = this.particles[j].y - particle.y;
                const distance = Math.sqrt(dx * dx + dy * dy);

                if (distance < this.connectionDistance) {
                    this.ctx.beginPath();
                    this.ctx.moveTo(particle.x, particle.y);
                    this.ctx.lineTo(this.particles[j].x, this.particles[j].y);
                    const opacity = (1 - distance / this.connectionDistance) * 0.3;
                    this.ctx.strokeStyle = `rgba(102, 126, 234, ${opacity})`;
                    this.ctx.lineWidth = 1;
                    this.ctx.stroke();
                }
            }
        });

        requestAnimationFrame(() => this.animate());
    }
}

// ========================================
// INITIALIZATION
// ========================================
document.addEventListener('DOMContentLoaded', () => {
    // Initialize particle system
    new ParticleSystem(elements.particlesCanvas);

    // Initialize event listeners
    initializeEventListeners();

    // Check existing session
    checkExistingSession();

    console.log('%c🚀 Order Management System', 'color: #667eea; font-size: 24px; font-weight: bold;');
    console.log('%cLogin Portal Ready', 'color: #764ba2; font-size: 14px;');
});

// ========================================
// EVENT LISTENERS
// ========================================
function initializeEventListeners() {
    // Form submission
    elements.loginForm.addEventListener('submit', handleFormSubmit);

    // Password visibility toggle
    elements.togglePassword.addEventListener('click', togglePasswordVisibility);

    // Clear alerts on input
    elements.identifierInput.addEventListener('input', clearAlert);
    elements.passwordInput.addEventListener('input', clearAlert);
}

// ========================================
// FORM SUBMISSION
// ========================================
async function handleFormSubmit(e) {
    e.preventDefault();
    clearAlert();

    const identifier = elements.identifierInput.value.trim();
    const password = elements.passwordInput.value;

    // Basic validation
    if (!identifier || !password) {
        showAlert('Please fill in all fields', 'error');
        elements.loginForm.classList.add('error-shake');
        setTimeout(() => elements.loginForm.classList.remove('error-shake'), 500);
        return;
    }

    // Show loading state
    setLoadingState(true);

    try {
        // Try user login first
        const response = await fetch(`${API_BASE_URL}/auth/user/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ identifier, password })
        });

        const data = await response.json();

        if (response.ok) {
            handleLoginSuccess(data, 'user');
        } else {
            // If user login fails, try admin login
            const adminResponse = await fetch(`${API_BASE_URL}/auth/admin/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ identifier, password })
            });

            const adminData = await adminResponse.json();

            if (adminResponse.ok) {
                handleLoginSuccess(adminData, 'admin');
            } else {
                handleLoginError(adminData);
            }
        }
    } catch (error) {
        console.error('Login error:', error);
        showAlert('Unable to connect to server. Please try again.', 'error');
    } finally {
        setLoadingState(false);
    }
}

function handleLoginSuccess(data, userType) {
    showAlert('Login successful! Redirecting...', 'success');

    // Store authentication data
    if (data.access_token) {
        localStorage.setItem('accessToken', data.access_token);
    }
    if (data.refresh_token) {
        localStorage.setItem('refreshToken', data.refresh_token);
    }

    // Determine user type from response
    const isSuperAdmin = data.user?.is_super_admin || data.user?.isSuperAdmin || false;
    const finalUserType = isSuperAdmin ? 'admin' : userType;

    localStorage.setItem('userType', finalUserType);
    localStorage.setItem('userEmail', data.user?.email || '');
    localStorage.setItem('userName', data.user?.name || '');
    localStorage.setItem('userData', JSON.stringify(data.user || {}));

    // Redirect based on user type
    setTimeout(() => {
        if (finalUserType === 'admin') {
            window.location.href = '/ui/admin/dashboard.html';
        } else {
            window.location.href = '/ui/user/dashboard.html';
        }
    }, 1000);
}

function handleLoginError(data) {
    const errorMessage = data.message || data.error || 'Invalid credentials. Please try again.';
    showAlert(errorMessage, 'error');

    elements.loginForm.classList.add('error-shake');
    setTimeout(() => elements.loginForm.classList.remove('error-shake'), 500);
}

// ========================================
// UI HELPERS
// ========================================
function setLoadingState(isLoading) {
    if (isLoading) {
        elements.submitBtn.classList.add('loading');
        elements.submitBtn.disabled = true;
    } else {
        elements.submitBtn.classList.remove('loading');
        elements.submitBtn.disabled = false;
    }
}

function togglePasswordVisibility() {
    const type = elements.passwordInput.type === 'password' ? 'text' : 'password';
    elements.passwordInput.type = type;

    const eyeIcon = elements.togglePassword.querySelector('.eye-icon');
    if (type === 'text') {
        eyeIcon.innerHTML = `
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
        `;
    } else {
        eyeIcon.innerHTML = `
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        `;
    }
}

// ========================================
// ALERT SYSTEM
// ========================================
function showAlert(message, type = 'info') {
    clearAlert();

    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;

    const icon = getAlertIcon(type);
    alertDiv.innerHTML = `
        <span style="font-size: 1.25rem;">${icon}</span>
        <span>${message}</span>
    `;

    elements.alertContainer.appendChild(alertDiv);

    if (type !== 'error') {
        setTimeout(clearAlert, 5000);
    }
}

function clearAlert() {
    elements.alertContainer.innerHTML = '';
}

function getAlertIcon(type) {
    const icons = {
        success: '✓',
        error: '✕',
        warning: '⚠',
        info: 'ℹ'
    };
    return icons[type] || icons.info;
}

// ========================================
// SESSION MANAGEMENT
// ========================================
function checkExistingSession() {
    const accessToken = localStorage.getItem('accessToken');
    const userType = localStorage.getItem('userType');

    if (accessToken && userType) {
        showAlert('You are already logged in. Redirecting...', 'info');
        setTimeout(() => {
            if (userType === 'admin') {
                window.location.href = '/ui/admin/dashboard.html';
            } else {
                window.location.href = '/ui/user/dashboard.html';
            }
        }, 1000);
    }
}

// ========================================
// UTILITY FUNCTIONS
// ========================================
function logout() {
    localStorage.clear();
    window.location.href = '/ui/index.html';
}

// Make logout available globally
window.logout = logout;
