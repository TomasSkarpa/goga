class Dashboard {
    constructor() {
        this.slideshow = document.getElementById('slideshow');
        this.layer1 = document.getElementById('layer1');
        this.layer2 = document.getElementById('layer2');
        this.galleryStrip = document.getElementById('galleryStrip');
        this.recentUploads = document.getElementById('recentUploads');
        this.uploadArea = document.getElementById('uploadArea');
        this.fileInput = document.getElementById('fileInput');
        this.uploadModal = document.getElementById('uploadModal');
        this.uploadProgress = document.getElementById('uploadProgress');
        this.configBtn = document.getElementById('configBtn');
        this.galleryBtn = document.getElementById('galleryBtn');
        this.configModal = document.getElementById('configModal');
        this.galleryModal = document.getElementById('galleryModal');
        this.galleryGrid = document.getElementById('galleryGrid');
        this.closeConfigBtn = document.getElementById('closeConfigBtn');
        this.closeGalleryBtn = document.getElementById('closeGalleryBtn');
        this.cancelConfigBtn = document.getElementById('cancelConfigBtn');
        this.saveConfigBtn = document.getElementById('saveConfigBtn');
        this.aiApiKeyInput = document.getElementById('aiApiKey');

        this.leftPanel = document.getElementById('leftPanel');
        this.bottomGallery = document.getElementById('bottomGallery');
        this.resizeHandle = document.getElementById('resizeHandle');

        this.galleryOffset = 0;
        this.isLoading = false;
        this.isResizing = false;
        this.panelWidth = 320;
        this.imageVersions = new Map(); // Smart cache invalidation
        
        this.images = [];
        this.slideshowImages = [];
        this.currentSlideIndex = 0;
        this.slideshowInterval = null;
        this.activeLayer = this.layer1;
        this.inactiveLayer = this.layer2;
        this.maxSlideshowImages = 8;
        
        this.initEventListeners();
        this.loadServerConfig();
        this.loadImages();
    }
    
    // Toast notification system
    showToast(message, type = 'success') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        toast.style.cssText = `
            position: fixed; top: 80px; right: 20px;
            background: rgba(0,0,0,0.9); color: white;
            padding: 12px 20px; border-radius: 8px;
            backdrop-filter: blur(10px); z-index: 1000;
            transform: translateX(400px);
            transition: transform 0.3s ease;
            border-left: 4px solid ${type === 'error' ? '#ef4444' : '#10b981'};
        `;
        document.body.appendChild(toast);
        
        setTimeout(() => toast.style.transform = 'translateX(0)', 100);
        setTimeout(() => {
            toast.style.transform = 'translateX(400px)';
            setTimeout(() => document.body.removeChild(toast), 300);
        }, 3000);
    }

    initEventListeners() {
        // Upload area
        this.uploadArea.addEventListener('click', () => this.fileInput.click());
        this.fileInput.addEventListener('change', (e) => this.handleFiles(e.target.files));

        // Drag and drop
        this.uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.uploadArea.classList.add('bg-blue-500/20', 'border-blue-400');
        });

        this.uploadArea.addEventListener('dragleave', () => {
            this.uploadArea.classList.remove('bg-blue-500/20', 'border-blue-400');
        });

        this.uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            this.uploadArea.classList.remove('bg-blue-500/20', 'border-blue-400');
            this.handleFiles(e.dataTransfer.files);
        });

        // Buttons
        this.configBtn.addEventListener('click', () => this.showConfig());
        this.galleryBtn.addEventListener('click', () => this.showFullGallery());
        
        // Config modal
        this.closeConfigBtn.addEventListener('click', () => this.hideConfig());
        this.cancelConfigBtn.addEventListener('click', () => this.hideConfig());
        this.saveConfigBtn.addEventListener('click', () => this.saveConfig());

        this.configModal.addEventListener('click', (e) => {
            if (e.target === this.configModal) this.hideConfig();
        });
        
        // Gallery modal
        this.closeGalleryBtn.addEventListener('click', () => this.hideGallery());
        this.galleryModal.addEventListener('click', (e) => {
            if (e.target === this.galleryModal) this.hideGallery();
        });
        
        // Infinite scroll
        this.galleryStrip.addEventListener('scroll', () => this.handleScroll());
        
        // Panel resize
        this.resizeHandle.addEventListener('mousedown', (e) => this.startResize(e));
        document.addEventListener('mousemove', (e) => this.handleResize(e));
        document.addEventListener('mouseup', () => this.stopResize());

        // Modal close
        this.uploadModal.addEventListener('click', (e) => {
            if (e.target === this.uploadModal) this.hideUploadModal();
        });
        
        // ESC key to close modals, Enter to save config
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideConfig();
                this.hideGallery();
                this.hideUploadModal();
            }
            if (e.key === 'Enter' && !this.configModal.classList.contains('hidden')) {
                this.saveConfig();
            }
        });
    }

    async loadImages() {
        try {
            const response = await fetch('/api/images');
            this.images = await response.json();
            
            // Create slideshow subset
            this.slideshowImages = this.images.slice(0, this.maxSlideshowImages);
            this.preloadSlideshowImages();
            
            this.renderGalleryStrip();
            this.renderRecentUploads();
            this.startSlideshow();
        } catch (error) {
            console.error('Failed to load images:', error);
            this.galleryStrip.innerHTML = '<div class="text-white/60 text-center py-8 px-4">Upload images to start</div>';
        }
    }

    renderGalleryStrip() {
        if (this.images.length === 0) {
            this.galleryStrip.innerHTML = '<div class="text-white/60 text-center py-4 px-4 flex-shrink-0">Upload images to start</div>';
            return;
        }

        // Sort by recent access (most recent first), then by creation date
        const sortedImages = this.getSortedImagesByAccess();
        
        this.galleryStrip.innerHTML = sortedImages.map(image => {
            const version = this.getImageVersion(image.id);
            return `
            <div class="flex-shrink-0 cursor-pointer group" onclick="window.dashboard.showImageDetail('${image.id}')">
                <div class="relative w-48 h-36 bg-gray-800 rounded-lg overflow-hidden">
                    <div class="absolute inset-0 bg-gradient-to-r from-gray-800 via-gray-700 to-gray-800 animate-pulse"></div>
                    <img src="/api/images/${image.id}/file?thumb=280&v=${version}" 
                         alt="${image.original_name}" 
                         class="absolute inset-0 w-full h-full object-cover rounded-lg shadow-soft group-hover:scale-105 transition-all duration-300 opacity-0" 
                         style="image-orientation: from-image;" loading="lazy"
                         onload="this.style.opacity='1'; this.previousElementSibling.style.display='none'">
                </div>
            </div>
            `;
        }).join('');
    }

    renderRecentUploads() {
        const recent = this.images.slice(0, 8);
        
        this.recentUploads.innerHTML = recent.map(image => {
            const version = this.getImageVersion(image.id);
            return `
            <div class="cursor-pointer group" onclick="window.dashboard.showImageDetail('${image.id}')">
                <div class="relative w-full aspect-square bg-gray-800 rounded-lg overflow-hidden">
                    <div class="absolute inset-0 bg-gradient-to-r from-gray-800 via-gray-700 to-gray-800 animate-pulse"></div>
                    <img src="/api/images/${image.id}/file?thumb=120&v=${version}" 
                         alt="${image.original_name}" 
                         class="absolute inset-0 w-full h-full object-cover rounded-lg shadow-inner-custom group-hover:scale-105 transition-all duration-300 opacity-0" 
                         style="image-orientation: from-image;" loading="lazy"
                         onload="this.style.opacity='1'; this.previousElementSibling.style.display='none'">
                </div>
            </div>
            `;
        }).join('');
    }

    preloadSlideshowImages() {
        this.slideshowImages.forEach(image => {
            const img = new Image();
            img.src = `/api/images/${image.id}/file`;
        });
    }

    startSlideshow() {
        if (this.slideshowImages.length === 0) return;
        
        // Set first image immediately
        const firstImage = this.slideshowImages[0];
        this.activeLayer.style.backgroundImage = `linear-gradient(rgba(0,0,0,0.3), rgba(0,0,0,0.3)), url('/api/images/${firstImage.id}/file')`;
        this.activeLayer.classList.add('active');
        
        if (this.slideshowImages.length > 1) {
            this.slideshowInterval = setInterval(() => {
                this.currentSlideIndex = (this.currentSlideIndex + 1) % this.slideshowImages.length;
                this.updateSlideshow();
            }, 8000);
        }
    }

    updateSlideshow() {
        if (this.slideshowImages.length === 0) return;
        
        const currentImage = this.slideshowImages[this.currentSlideIndex];
        
        // Set new image on inactive layer
        this.inactiveLayer.style.backgroundImage = `linear-gradient(rgba(0,0,0,0.3), rgba(0,0,0,0.3)), url('/api/images/${currentImage.id}/file')`;
        
        // Crossfade transition
        this.inactiveLayer.classList.add('active');
        this.activeLayer.classList.remove('active');
        
        // Swap layers
        [this.activeLayer, this.inactiveLayer] = [this.inactiveLayer, this.activeLayer];
    }

    async handleFiles(files) {
        const fileArray = Array.from(files);
        this.showUploadModal();
        
        for (let i = 0; i < fileArray.length; i++) {
            const file = fileArray[i];
            await this.uploadFile(file, i + 1, fileArray.length);
        }
        
        this.hideUploadModal();
        this.loadImages();
    }

    async uploadFile(file, current, total) {
        const formData = new FormData();
        formData.append('image', file);

        const progressDiv = document.createElement('div');
        progressDiv.className = 'mb-3';
        progressDiv.innerHTML = `
            <div class="flex justify-between text-white/90 text-sm mb-2">
                <span class="truncate">${file.name}</span>
                <span>${current}/${total}</span>
            </div>
            <div class="w-full bg-white/20 rounded-full h-2">
                <div class="bg-blue-500 h-2 rounded-full transition-all duration-300" style="width: 0%"></div>
            </div>
        `;
        this.uploadProgress.appendChild(progressDiv);

        const progressBar = progressDiv.querySelector('.bg-blue-500');

        try {
            const response = await fetch('/api/images/upload', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                progressBar.style.width = '100%';
                progressBar.classList.remove('bg-blue-500');
                progressBar.classList.add('bg-green-500');
            } else {
                throw new Error('Upload failed');
            }
        } catch (error) {
            console.error('Upload failed:', error);
            progressBar.classList.remove('bg-blue-500');
            progressBar.classList.add('bg-red-500');
            progressBar.style.width = '100%';
        }
    }

    showUploadModal() {
        this.uploadModal.classList.remove('hidden');
        this.uploadProgress.innerHTML = '';
    }

    hideUploadModal() {
        this.uploadModal.classList.add('hidden');
        this.uploadProgress.innerHTML = '';
        this.fileInput.value = '';
    }

    getSortedImagesByAccess() {
        const recentAccess = JSON.parse(localStorage.getItem('recentImageAccess') || '{}');
        
        return this.images.sort((a, b) => {
            const aAccess = recentAccess[a.id] || 0;
            const bAccess = recentAccess[b.id] || 0;
            
            if (aAccess !== bAccess) {
                return bAccess - aAccess; // Most recent first
            }
            
            return new Date(b.created_at) - new Date(a.created_at); // Then by creation date
        });
    }
    
    trackImageAccess(imageId) {
        const recentAccess = JSON.parse(localStorage.getItem('recentImageAccess') || '{}');
        recentAccess[imageId] = Date.now();
        localStorage.setItem('recentImageAccess', JSON.stringify(recentAccess));
    }
    
    handleScroll() {
        const { scrollLeft, scrollWidth, clientWidth } = this.galleryStrip;
        
        // Remove fade mask when at the end
        if (scrollLeft + clientWidth >= scrollWidth - 50) {
            this.galleryStrip.classList.remove('fade-mask');
        } else {
            this.galleryStrip.classList.add('fade-mask');
        }
        
        if (scrollLeft + clientWidth >= scrollWidth - 100 && !this.isLoading) {
            this.loadMoreImages();
        }
    }

    showImageDetail(imageId) {
        this.trackImageAccess(imageId);
        window.location.href = `/image/${imageId}`;
    }

    async showConfig() {
        await this.loadServerConfig();
        this.loadConfig();
        this.configModal.classList.remove('hidden');
    }
    
    hideConfig() {
        this.configModal.classList.add('hidden');
    }
    
    async loadServerConfig() {
        try {
            const response = await fetch('/api/config');
            if (response.ok) {
                const config = await response.json();
                localStorage.setItem('gogaConfig', JSON.stringify(config));
                // Update placeholder based on server state
                if (config.hasApiKey && this.aiApiKeyInput) {
                    this.aiApiKeyInput.placeholder = 'API key configured (hidden)';
                }
            }
        } catch (error) {
            console.error('Failed to load server config:', error);
        }
    }
    
    loadConfig() {
        const config = JSON.parse(localStorage.getItem('gogaConfig') || '{}');
        // Don't populate API key field for security - user must re-enter
        this.aiApiKeyInput.value = '';
        this.aiApiKeyInput.placeholder = config.hasApiKey ? 'API key configured (hidden)' : 'Enter your AI Studio API key';
    }
    
    saveConfig() {
        const apiKey = this.aiApiKeyInput.value.trim();
        
        // Only save if API key was entered
        if (!apiKey) {
            this.hideConfig();
            return;
        }
        
        const config = { aiApiKey: apiKey };
        
        // Check if key already exists
        const existingConfig = JSON.parse(localStorage.getItem('gogaConfig') || '{}');
        const isOverwrite = existingConfig.hasApiKey;
        
        // Send to server
        fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ aiApiKey: apiKey })
        }).then(response => {
            if (response.ok) {
                // Clear input for security
                this.aiApiKeyInput.value = '';
                // Update localStorage to show key is configured
                localStorage.setItem('gogaConfig', JSON.stringify({hasApiKey: true}));
                this.aiApiKeyInput.placeholder = 'API key configured (hidden)';
                this.hideConfig();
                this.showToast(isOverwrite ? 'API key overwritten!' : 'API key saved securely!');
            } else {
                throw new Error('Server error');
            }
        }).catch(err => {
            console.error('Failed to save config:', err);
            this.showToast('Failed to save configuration', 'error');
        });
    }

    


    showFullGallery() {
        this.renderGalleryGrid();
        this.galleryModal.classList.remove('hidden');
    }
    
    hideGallery() {
        this.galleryModal.classList.add('hidden');
    }
    
    renderGalleryGrid() {
        if (this.images.length === 0) {
            this.galleryGrid.innerHTML = '<div class="col-span-full text-white/60 text-center py-8">No images uploaded yet</div>';
            return;
        }

        const sortedImages = this.getSortedImagesByAccess();
        
        this.galleryGrid.innerHTML = sortedImages.map(image => {
            const version = this.getImageVersion(image.id);
            return `
            <div class="cursor-pointer group" onclick="window.dashboard.showImageDetail('${image.id}')">
                <div class="relative aspect-square bg-gray-800 rounded-lg overflow-hidden">
                    <div class="absolute inset-0 bg-gradient-to-r from-gray-800 via-gray-700 to-gray-800 animate-pulse"></div>
                    <img src="/api/images/${image.id}/file?thumb=280&v=${version}" 
                         alt="${image.original_name}" 
                         class="absolute inset-0 w-full h-full object-cover rounded-lg group-hover:scale-105 transition-all duration-300 opacity-0" 
                         style="image-orientation: from-image;" loading="lazy"
                         onload="this.style.opacity='1'; this.previousElementSibling.style.display='none'">
                    <div class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-all duration-300 rounded-lg"></div>
                    <div class="absolute bottom-2 left-2 right-2 text-white text-xs truncate opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                        ${image.original_name}
                    </div>
                </div>
            </div>
            `;
        }).join('');
    }

    startResize(e) {
        this.isResizing = true;
        e.preventDefault();
    }
    
    handleResize(e) {
        if (!this.isResizing) return;
        
        const newWidth = Math.max(280, Math.min(500, e.clientX));
        this.panelWidth = newWidth;
        
        this.leftPanel.style.width = `${newWidth}px`;
        this.bottomGallery.style.left = `calc(${newWidth}px + 2rem)`;
    }
    
    stopResize() {
        this.isResizing = false;
    }

    async loadMoreImages() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        // In a real app, this would fetch more images from the API
        // For now, we'll just re-render existing images
        setTimeout(() => {
            this.isLoading = false;
        }, 500);
    }
    
    // Smart cache invalidation - only refresh edited images
    getImageVersion(imageId) {
        if (!this.imageVersions.has(imageId)) {
            this.imageVersions.set(imageId, 1);
        }
        return this.imageVersions.get(imageId);
    }
    
    // Increment version when image is edited
    invalidateImageCache(imageId) {
        const currentVersion = this.getImageVersion(imageId);
        this.imageVersions.set(imageId, currentVersion + 1);
    }
    
    // Refresh gallery thumbnails
    refreshGallery() {
        this.renderGalleryStrip();
        this.renderRecentUploads();
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new Dashboard();
    
    // Listen for image edit events from other tabs/windows
    window.addEventListener('storage', (e) => {
        if (e.key === 'imageEdited') {
            const imageId = e.newValue;
            window.dashboard.invalidateImageCache(imageId);
            window.dashboard.refreshGallery();
        }
    });
    
    // Listen for page visibility changes
    document.addEventListener('visibilitychange', () => {
        if (!document.hidden) {
            // Only refresh if we detect potential changes
            const lastRefresh = localStorage.getItem('lastGalleryRefresh');
            const now = Date.now();
            if (!lastRefresh || now - parseInt(lastRefresh) > 30000) { // 30 seconds
                setTimeout(() => window.dashboard.refreshGallery(), 100);
                localStorage.setItem('lastGalleryRefresh', now.toString());
            }
        }
    });
});