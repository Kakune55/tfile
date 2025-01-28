// File Manager Web UI - Main JavaScript
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

// 全局状态
let currentPath = '';
const uploads = new Map(); // 用于跟踪上传任务

/* 初始化应用 */
function initializeApp() {
    // 初始化路径
    const urlParams = new URLSearchParams(window.location.search);
    const initialPath = urlParams.get('path') || '';
    currentPath = normalizePath(initialPath);
    
    // 初始加载文件
    loadFiles(currentPath);
    
    // 注册全局事件
    registerGlobalEvents();
}

/* 注册全局事件 */
function registerGlobalEvents() {
    // 键盘导航 (Alt + Backspace 返回上级)
    document.addEventListener('keyup', (e) => {
        if (e.altKey && e.key === 'Backspace') goUp();
    });

    // 文件输入变化时自动显示文件名
    document.getElementById('fileInput').addEventListener('change', showSelectedFiles);
}

/* 文件操作核心函数 */
async function loadFiles(path) {
    try {
        currentPath = normalizePath(path);
        updateBrowserURL();
        
        const query = currentPath ? `?path=${encodeURIComponent(currentPath)}` : '';
        const response = await fetch(`/api/list${query}`);
        
        if (!response.ok) throw new Error('Failed to load files');
        
        const files = await response.json();
        renderFileList(files);
        updatePathDisplay();
    } catch (error) {
        showError(error.message);
    }
}

/* 文件列表渲染 */
function renderFileList(files) {
    const fileList = document.getElementById('fileList');
    let html = `
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Size</th>
                    <th>Modified</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>`;
    
    files.forEach(file => {
        html += `
            <tr>
                <td>
                    ${file.isDir ? 
                        `<span class="dir-link" onclick="navigateTo('${encodeURIComponent(file.path)}')">
                            📁 ${escapeHtml(file.name)}
                        </span>` : 
                        `📄 ${escapeHtml(file.name)}`}
                </td>
                <td>${file.isDir ? '-' : formatSize(file.size)}</td>
                <td>${new Date(file.modTime).toLocaleString()}</td>
                <td>
                    <div class="action-buttons">
                        ${!file.isDir ? 
                            `<button class="btn btn-primary" 
                                     onclick="downloadFile('${encodeURIComponent(file.path)}')">
                                ⬇️ Download
                            </button>` : ''}
                        <button class="btn btn-secondary" 
                                onclick="promptRename('${encodeURIComponent(file.path)}')">
                            ✏️ Rename
                        </button>
                        <button class="btn btn-danger" 
                                onclick="confirmDelete('${encodeURIComponent(file.path)}')">
                            🗑️ Delete
                        </button>
                    </div>
                </td>
            </tr>`;
    });
    
    html += `</tbody></table>`;
    fileList.innerHTML = html;
}

/* 路径导航功能 */
function updatePathDisplay() {
const parts = currentPath.replace(/\\/g, '/').split('/').filter(p => p);
    let pathHtml = `<span class="path-segment" onclick="navigateTo('')">🏠 Home</span>`;
    
    let accumulated = [];
    parts.forEach((part, index) => {
        accumulated.push(part);
        const fullPath = accumulated.join('/');
        pathHtml += `
            <span class="path-separator">/</span>
            <span class="path-segment" 
                  onclick="navigateTo('${encodeURIComponent(fullPath)}')"
                  title="${escapeHtml(fullPath)}">
                ${escapeHtml(part)}
            </span>`;
    });
    
    document.getElementById('pathDisplay').innerHTML = pathHtml;
    document.getElementById('upButton').disabled = currentPath === '';
}

function navigateTo(path) {
    currentPath = normalizePath(path);
    loadFiles(currentPath);
}

function goUp() {
    const parts = currentPath.split('/').filter(p => p);
    if (parts.length > 0) {
        parts.pop();
        currentPath = parts.join('/');
        loadFiles(currentPath);
    }
}

/* 文件上传功能 */
function startUpload() {
    const input = document.getElementById('fileInput');
    const files = input.files;
    
    if (files.length === 0) {
        showError('Please select files to upload');
        return;
    }

    Array.from(files).forEach(file => {
        const uploadId = generateUploadId();
        createProgressItem(uploadId, file);
        uploadFile(uploadId, file);
    });
    
    input.value = ''; // 清空选择
}

function generateUploadId() {
    return `upload-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function createProgressItem(uploadId, file) {
    const container = document.getElementById('uploadProgress');
    
    const item = document.createElement('div');
    item.className = 'upload-item';
    item.id = uploadId;
    item.innerHTML = `
        <div class="file-info">
            <span class="file-name">📄 ${escapeHtml(file.name)}</span>
            <button class="btn btn-danger btn-sm" onclick="cancelUpload('${uploadId}')">Cancel</button>
        </div>
        <div class="progress-bar">
            <div class="progress-fill" style="width: 0%"></div>
        </div>
        <div class="upload-stats">
            <span class="progress-percent">0%</span>
            <span class="upload-speed">0 KB/s</span>
            <span class="upload-status">Waiting...</span>
        </div>
    `;
    
    container.prepend(item);
}

function uploadFile(uploadId, file) {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    formData.append('files', file);
    formData.append('path', currentPath);

    let startTime = Date.now();
    let lastLoaded = 0;
    let lastTime = startTime;
    
    uploads.set(uploadId, { xhr, startTime, file });

    xhr.upload.addEventListener('progress', (e) => {
        if (!e.lengthComputable) return;
        
        const currentTime = Date.now();
        const deltaTime = (currentTime - lastTime) / 1000;
        const deltaLoaded = e.loaded - lastLoaded;
        const speed = deltaLoaded / deltaTime;
        
        updateProgress(uploadId, {
            progress: (e.loaded / e.total) * 100,
            speed: speed,
            loaded: e.loaded,
            total: e.total
        });

        lastLoaded = e.loaded;
        lastTime = currentTime;
    });

    xhr.upload.addEventListener('error', () => {
        updateProgress(uploadId, { 
            status: 'Error', 
            error: 'Network error' 
        });
        cleanupUpload(uploadId);
    });

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4) {
            if (xhr.status === 201) {
                updateProgress(uploadId, { 
                    progress: 100, 
                    status: 'Completed' 
                });
                //上传成功后刷新列表
                loadFiles(currentPath);
                setTimeout(() => cleanupUpload(uploadId), 2000);
            } else {
                updateProgress(uploadId, { 
                    status: 'Error',
                    error: xhr.statusText || 'Upload failed'
                });
            }
            cleanupUpload(uploadId);
        }
    };

    xhr.open('POST', '/api/upload', true);
    xhr.send(formData);
}

function updateProgress(uploadId, data) {
    const item = document.getElementById(uploadId);
    if (!item) return;

    // 更新进度条
    if (data.progress !== undefined) {
        item.querySelector('.progress-fill').style.width = `${data.progress}%`;
        item.querySelector('.progress-percent').textContent = 
            `${data.progress.toFixed(1)}%`;
    }

    // 更新速度
    if (data.speed !== undefined) {
        item.querySelector('.upload-speed').textContent = formatSpeed(data.speed);
    }

    // 更新状态
    if (data.status) {
        const statusElement = item.querySelector('.upload-status');
        statusElement.textContent = data.status;
        statusElement.style.color = 
            data.status === 'Completed' ? '#28a745' : 
            data.status === 'Error' ? '#dc3545' : '#666';
    }

    if (data.error) showError(data.error);
}

function cancelUpload(uploadId) {
    const upload = uploads.get(uploadId);
    if (upload) {
        upload.xhr.abort();
        cleanupUpload(uploadId);
    }
}

function cleanupUpload(uploadId) {
    document.getElementById(uploadId)?.remove();
    uploads.delete(uploadId);
}

/* 其他文件操作 */
async function createFolder() {
    const folderName = document.getElementById('newFolderName').value.trim();
    if (!folderName) {
        showError('Please enter a folder name');
        return;
    }

    try {
        const response = await fetch('/api/mkdir', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name: folderName,
                path: currentPath
            })
        });

        if (!response.ok) throw new Error('Failed to create folder');
        
        document.getElementById('newFolderName').value = '';
        loadFiles(currentPath);
    } catch (error) {
        showError(error.message);
    }
}

function downloadFile(filePath) {
    try {
        const encodedPath = encodeURIComponent(encodeURIComponent(filePath));
        const link = document.createElement('a');
        link.href = `/api/download/${encodedPath}`;
        link.style.display = 'none';
        link.download = decodeURIComponent(filePath).split('/').pop();
        
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(link.href);
    } catch (error) {
        showError(`Download failed: ${error.message}`);
    }
}

function promptRename(oldPath) {
    const decodedPath = decodeURIComponent(oldPath);
    const oldName = decodedPath.split('/').pop();
    const newName = prompt('Enter new name:', oldName);
    
    if (!newName || newName === oldName) return;

    const newPath = filepathJoin(
        currentPath, 
        decodedPath.split('/').slice(0, -1).join('/'), 
        newName
    );

    performRename(decodedPath, newPath);
}

async function performRename(oldPath, newPath) {
    try {
        const response = await fetch('/api/rename', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ old: oldPath, new: newPath })
        });

        if (!response.ok) throw new Error('Rename failed');
        loadFiles(currentPath);
    } catch (error) {
        showError(error.message);
    }
}

function confirmDelete(filePath) {
    const decodedPath = decodeURIComponent(filePath);
    const fileName = decodedPath.split('/').pop();
    
    if (!confirm(`Delete "${fileName}" permanently?`)) return;
    
    performDelete(decodedPath);
}

async function performDelete(filePath) {
    try {
        const response = await fetch('/api/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: filePath })
        });

        if (!response.ok) throw new Error('Delete failed');
        loadFiles(currentPath);
    } catch (error) {
        showError(error.message);
    }
}

/* 辅助函数 */
function normalizePath(path) {
    return decodeURIComponent(path)
        .split('/')
        .filter(p => p)
        .join('/');
}

function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function formatSpeed(bytesPerSecond) {
    if (bytesPerSecond === 0) return '0 B/s';
    const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
    const i = Math.floor(Math.log(bytesPerSecond) / Math.log(1024));
    return `${(bytesPerSecond / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function escapeHtml(str) {
    return str.replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function filepathJoin(...parts) {
    return parts.filter(p => p).join('/').replace(/\/+/g, '/');
}

function updateBrowserURL() {
    const newUrl = window.location.pathname + 
        (currentPath ? `?path=${encodeURIComponent(currentPath)}` : '');
    window.history.replaceState({}, '', newUrl);
}

function showError(message) {
    const errorToast = document.createElement('div');
    errorToast.className = 'error-toast';
    errorToast.textContent = message;
    
    document.body.appendChild(errorToast);
    setTimeout(() => errorToast.remove(), 3000);
}

function showSelectedFiles() {
    const input = document.getElementById('fileInput');
    // 可以在这里添加选中文件的预览功能
}