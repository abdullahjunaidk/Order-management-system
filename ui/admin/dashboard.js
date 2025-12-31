// ========================================
// CONFIGURATION
// ========================================
const API_BASE_URL = 'http://localhost:8080/api/v1';

// ========================================
// STATE
// ========================================
let currentPage = 'overview';
let accessToken = localStorage.getItem('accessToken');
let userData = JSON.parse(localStorage.getItem('userData') || '{}');

// ========================================
// INITIALIZATION
// ========================================
document.addEventListener('DOMContentLoaded', () => {
    // Check authentication
    if (!accessToken) {
        window.location.href = '../index.html';
        return;
    }

    // Initialize UI
    initializeUI();

    // Load initial page
    loadPage('overview');

    console.log('%c🚀 Admin Dashboard Ready', 'color: #667eea; font-size: 18px; font-weight: bold;');
});

// ========================================
// UI INITIALIZATION
// ========================================
function initializeUI() {
    // Set user name
    const userName = userData.name || userData.username || 'Admin';
    document.getElementById('userName').textContent = userName;

    // Navigation items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const page = item.dataset.page;
            loadPage(page);
        });
    });

    // Mobile menu
    document.getElementById('mobileMenuBtn').addEventListener('click', () => {
        document.getElementById('sidebar').classList.toggle('active');
    });
}

// ========================================
// PAGE LOADING
// ========================================
function loadPage(page) {
    currentPage = page;

    // Update active nav item
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
        if (item.dataset.page === page) {
            item.classList.add('active');
        }
    });

    // Update page title
    const titles = {
        overview: 'Overview',
        companies: 'Companies',
        users: 'Users',
        products: 'Products',
        inventory: 'Inventory',
        orders: 'Orders',
        roles: 'Roles'
    };
    document.getElementById('pageTitle').textContent = titles[page] || 'Dashboard';

    // Load page content
    const content = document.getElementById('content');
    content.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

    setTimeout(() => {
        switch (page) {
            case 'overview':
                loadOverview();
                break;
            case 'companies':
                loadCompanies();
                break;
            case 'users':
                loadUsers();
                break;
            case 'products':
                loadProducts();
                break;
            case 'inventory':
                loadInventory();
                break;
            case 'orders':
                loadOrders();
                break;
            case 'roles':
                loadRoles();
                break;
            default:
                loadOverview();
        }
    }, 300);
}

// ========================================
// OVERVIEW PAGE
// ========================================
function loadOverview() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-header">
                    <span class="stat-title">Total Companies</span>
                    <div class="stat-icon primary">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                        </svg>
                    </div>
                </div>
                <div class="stat-value" id="totalCompanies">-</div>
                <div class="stat-change positive">
                    <span>↑ 12%</span>
                    <span>from last month</span>
                </div>
            </div>
            
            <div class="stat-card">
                <div class="stat-header">
                    <span class="stat-title">Total Users</span>
                    <div class="stat-icon success">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                    </div>
                </div>
                <div class="stat-value" id="totalUsers">-</div>
                <div class="stat-change positive">
                    <span>↑ 8%</span>
                    <span>from last month</span>
                </div>
            </div>
            
            <div class="stat-card">
                <div class="stat-header">
                    <span class="stat-title">Total Products</span>
                    <div class="stat-icon warning">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                        </svg>
                    </div>
                </div>
                <div class="stat-value" id="totalProducts">-</div>
                <div class="stat-change positive">
                    <span>↑ 15%</span>
                    <span>from last month</span>
                </div>
            </div>
            
            <div class="stat-card">
                <div class="stat-header">
                    <span class="stat-title">Total Orders</span>
                    <div class="stat-icon error">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                        </svg>
                    </div>
                </div>
                <div class="stat-value" id="totalOrders">-</div>
                <div class="stat-change negative">
                    <span>↓ 3%</span>
                    <span>from last month</span>
                </div>
            </div>
        </div>
        
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Recent Activity</h2>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Activity</th>
                            <th>User</th>
                            <th>Time</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody id="recentActivity">
                        <tr>
                            <td colspan="4" class="empty-state">Loading...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;

    // Load stats
    loadStats();
}

async function loadStats() {
    try {
        // These are placeholder values - you'll need to implement actual API calls
        document.getElementById('totalCompanies').textContent = '12';
        document.getElementById('totalUsers').textContent = '48';
        document.getElementById('totalProducts').textContent = '156';
        document.getElementById('totalOrders').textContent = '89';
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

// ========================================
// COMPANIES PAGE
// ========================================
function loadCompanies() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Companies</h2>
                <div class="table-actions">
                    <button class="btn btn-primary" onclick="createCompany()">
                        + Add Company
                    </button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Description</th>
                            <th>Status</th>
                            <th>Created</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="companiesTable">
                        <tr>
                            <td colspan="5" class="empty-state">Loading companies...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;

    fetchCompanies();
}

async function fetchCompanies() {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/companies`, {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            displayCompanies(data.companies || []);
        } else {
            document.getElementById('companiesTable').innerHTML = `
                <tr><td colspan="5" class="empty-state">Failed to load companies</td></tr>
            `;
        }
    } catch (error) {
        console.error('Error fetching companies:', error);
        document.getElementById('companiesTable').innerHTML = `
            <tr><td colspan="5" class="empty-state">Error loading companies</td></tr>
        `;
    }
}

function displayCompanies(companies) {
    const tbody = document.getElementById('companiesTable');

    if (companies.length === 0) {
        tbody.innerHTML = `
            <tr><td colspan="5" class="empty-state">No companies found</td></tr>
        `;
        return;
    }

    tbody.innerHTML = companies.map(company => `
        <tr>
            <td><strong>${company.name}</strong></td>
            <td>${company.description || '-'}</td>
            <td>
                <span style="color: ${company.isActive ? 'var(--success)' : 'var(--error)'}">
                    ${company.isActive ? 'Active' : 'Inactive'}
                </span>
            </td>
            <td>${new Date(company.createdAt).toLocaleDateString()}</td>
            <td>
                <button class="btn btn-secondary" onclick="editCompany('${company.id}')">Edit</button>
            </td>
        </tr>
    `).join('');
}

// ========================================
// USERS PAGE
// ========================================
function loadUsers() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Users</h2>
                <div class="table-actions">
                    <button class="btn btn-primary" onclick="createUser()">
                        + Add User
                    </button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Username</th>
                            <th>Email</th>
                            <th>Phone</th>
                            <th>Status</th>
                            <th>Role</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="usersTable">
                        <tr>
                            <td colspan="7" class="empty-state">User management coming soon...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;
}

// ========================================
// PRODUCTS PAGE
// ========================================
function loadProducts() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Products</h2>
                <div class="table-actions">
                    <button class="btn btn-primary" onclick="createProduct()">
                        + Add Product
                    </button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>SKU</th>
                            <th>Price</th>
                            <th>Stock</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="productsTable">
                        <tr>
                            <td colspan="6" class="empty-state">Product management coming soon...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;
}

// ========================================
// INVENTORY PAGE
// ========================================
function loadInventory() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Inventory</h2>
                <div class="table-actions">
                    <button class="btn btn-primary" onclick="updateInventory()">
                        Update Stock
                    </button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Product</th>
                            <th>SKU</th>
                            <th>Available</th>
                            <th>Reserved</th>
                            <th>Total</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="inventoryTable">
                        <tr>
                            <td colspan="7" class="empty-state">Inventory management coming soon...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;
}

// ========================================
// ORDERS PAGE
// ========================================
function loadOrders() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Orders</h2>
                <div class="table-actions">
                    <button class="btn btn-secondary">Export</button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Order ID</th>
                            <th>Customer</th>
                            <th>Products</th>
                            <th>Total</th>
                            <th>Status</th>
                            <th>Date</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="ordersTable">
                        <tr>
                            <td colspan="7" class="empty-state">Order management coming soon...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;
}

// ========================================
// ROLES PAGE
// ========================================
function loadRoles() {
    const content = document.getElementById('content');
    content.innerHTML = `
        <div class="table-container">
            <div class="table-header">
                <h2 class="table-title">Roles & Permissions</h2>
                <div class="table-actions">
                    <button class="btn btn-primary" onclick="createRole()">
                        + Add Role
                    </button>
                </div>
            </div>
            <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Role Name</th>
                            <th>Description</th>
                            <th>Permissions</th>
                            <th>Users</th>
                            <th>Created</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="rolesTable">
                        <tr>
                            <td colspan="6" class="empty-state">Loading roles...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    `;

    fetchRoles();
}

async function fetchRoles() {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/admin/roles`, {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            displayRoles(data.roles || []);
        } else {
            document.getElementById('rolesTable').innerHTML = `
                <tr><td colspan="6" class="empty-state">Failed to load roles</td></tr>
            `;
        }
    } catch (error) {
        console.error('Error fetching roles:', error);
        document.getElementById('rolesTable').innerHTML = `
            <tr><td colspan="6" class="empty-state">Error loading roles</td></tr>
        `;
    }
}

function displayRoles(roles) {
    const tbody = document.getElementById('rolesTable');

    if (roles.length === 0) {
        tbody.innerHTML = `
            <tr><td colspan="6" class="empty-state">No roles found</td></tr>
        `;
        return;
    }

    tbody.innerHTML = roles.map(role => `
        <tr>
            <td><strong>${role.name}</strong></td>
            <td>${role.description || '-'}</td>
            <td>-</td>
            <td>-</td>
            <td>${new Date(role.createdAt).toLocaleDateString()}</td>
            <td>
                <button class="btn btn-secondary" onclick="editRole('${role.id}')">Edit</button>
            </td>
        </tr>
    `).join('');
}

// ========================================
// PLACEHOLDER FUNCTIONS
// ========================================
function createCompany() {
    alert('Create company functionality coming soon!');
}

function editCompany(id) {
    alert(`Edit company ${id} - Coming soon!`);
}

function createUser() {
    alert('Create user functionality coming soon!');
}

function createProduct() {
    alert('Create product functionality coming soon!');
}

function updateInventory() {
    alert('Update inventory functionality coming soon!');
}

function createRole() {
    alert('Create role functionality coming soon!');
}

function editRole(id) {
    alert(`Edit role ${id} - Coming soon!`);
}

// ========================================
// LOGOUT
// ========================================
function logout() {
    localStorage.clear();
    window.location.href = '../index.html';
}

// Make logout available globally
window.logout = logout;
