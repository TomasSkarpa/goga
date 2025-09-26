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
        this.closeConfigBtn = document.getElementById('closeConfigBtn');
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
        
        this.galleryStrip.innerHTML = sortedImages.map(image => `
            <div class="flex-shrink-0 cursor-pointer group" onclick="window.dashboard.showImageDetail('${image.id}')">
                <img src="/api/images/${image.id}/file?thumb=280" 
                     alt="${image.original_name}" 
                     class="w-48 h-36 object-cover rounded-lg shadow-soft group-hover:scale-105 transition-transform" style="image-orientation: from-image;">
            </div>
        `).join('');
    }

    renderRecentUploads() {
        const recent = this.images.slice(0, 8);
        
        this.recentUploads.innerHTML = recent.map(image => `
            <div class="cursor-pointer group" onclick="window.dashboard.showImageDetail('${image.id}')">
                <img src="/api/images/${image.id}/file?thumb=120" 
                     alt="${image.original_name}" 
                     class="w-full aspect-square object-cover rounded-lg shadow-inner-custom group-hover:scale-105 transition-transform" style="image-orientation: from-image;">
            </div>
        `).join('');
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
                alert(isOverwrite ? 'API key overwritten!' : 'API key saved securely!');
            } else {
                throw new Error('Server error');
            }
        }).catch(err => {
            console.error('Failed to save config:', err);
            alert('Failed to save configuration');
        });
    }

    


    showFullGallery() {
        // Toggle to full gallery view
        alert('Full gallery view - Coming soon!');
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
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new Dashboard();
});