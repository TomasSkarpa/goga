class Dashboard {
    constructor() {
        this.gallery = document.getElementById('gallery');
        this.uploadBtn = document.getElementById('uploadBtn');
        this.uploadModal = document.getElementById('uploadModal');
        this.dropZone = document.getElementById('dropZone');
        this.modalFileInput = document.getElementById('modalFileInput');
        this.uploadProgress = document.getElementById('uploadProgress');
        
        this.initEventListeners();
        this.loadImages();
    }

    initEventListeners() {
        this.uploadBtn.addEventListener('click', () => this.showUploadModal());
        
        // Modal close
        this.uploadModal.querySelector('.close').addEventListener('click', () => this.hideUploadModal());
        this.uploadModal.addEventListener('click', (e) => {
            if (e.target === this.uploadModal) this.hideUploadModal();
        });

        // File input
        this.modalFileInput.addEventListener('change', (e) => this.handleFiles(e.target.files));

        // Drag and drop
        this.dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.dropZone.classList.add('border-blue-400', 'bg-blue-50');
        });

        this.dropZone.addEventListener('dragleave', () => {
            this.dropZone.classList.remove('border-blue-400', 'bg-blue-50');
        });

        this.dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            this.dropZone.classList.remove('border-blue-400', 'bg-blue-50');
            this.handleFiles(e.dataTransfer.files);
        });
    }

    showUploadModal() {
        this.uploadModal.classList.remove('hidden');
    }

    hideUploadModal() {
        this.uploadModal.classList.add('hidden');
        this.uploadProgress.innerHTML = '';
        this.modalFileInput.value = '';
    }

    async loadImages() {
        try {
            const response = await fetch('/api/images');
            const images = await response.json();
            this.renderImages(images);
        } catch (error) {
            console.error('Failed to load images:', error);
            this.gallery.innerHTML = '<div class="col-span-full text-center py-12 text-red-500">Failed to load images</div>';
        }
    }

    renderImages(images) {
        if (images.length === 0) {
            this.gallery.innerHTML = '<div class="col-span-full text-center py-12 text-gray-500">No images uploaded yet</div>';
            return;
        }

        this.gallery.innerHTML = images.map(image => `
            <div class="bg-white rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow cursor-pointer" onclick="window.location.href='/image/${image.id}'">
                <img src="/api/images/${image.id}/file?thumb=250" alt="${image.original_name}" class="w-full h-48 object-cover">
                <div class="p-4">
                    <h3 class="font-medium text-gray-900 truncate">${image.original_name}</h3>
                    <p class="text-sm text-gray-500 mt-1">${this.formatFileSize(image.size)} • ${image.format.toUpperCase()}</p>
                    <p class="text-xs text-gray-400 mt-1">${new Date(image.created_at).toLocaleDateString()}</p>
                </div>
            </div>
        `).join('');
    }

    async handleFiles(files) {
        const fileArray = Array.from(files);
        
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
        progressDiv.className = 'mb-2';
        progressDiv.innerHTML = `
            <div class="flex justify-between text-sm text-gray-600 mb-1">
                <span>${file.name}</span>
                <span>${current}/${total}</span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2">
                <div class="bg-blue-600 h-2 rounded-full transition-all duration-300" style="width: 0%"></div>
            </div>
        `;
        this.uploadProgress.appendChild(progressDiv);

        const progressBar = progressDiv.querySelector('.bg-blue-600');

        try {
            const response = await fetch('/api/images/upload', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                progressBar.style.width = '100%';
                progressBar.classList.remove('bg-blue-600');
                progressBar.classList.add('bg-green-500');
            } else {
                throw new Error('Upload failed');
            }
        } catch (error) {
            console.error('Upload failed:', error);
            progressBar.classList.remove('bg-blue-600');
            progressBar.classList.add('bg-red-500');
            progressBar.style.width = '100%';
        }
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new Dashboard();
});