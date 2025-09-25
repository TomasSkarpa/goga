class ImageDetail {
    constructor(imageID) {
        this.imageID = imageID;
        this.mainImage = document.getElementById('mainImage');
        this.imageName = document.getElementById('imageName');
        this.imageDimensions = document.getElementById('imageDimensions');
        this.imageSize = document.getElementById('imageSize');
        this.imageFormat = document.getElementById('imageFormat');
        this.imageCreated = document.getElementById('imageCreated');
        this.convertFormat = document.getElementById('convertFormat');
        this.convertBtn = document.getElementById('convertBtn');
        this.deleteBtn = document.getElementById('deleteBtn');
        
        this.initEventListeners();
        this.loadImageDetails();
    }

    initEventListeners() {
        this.convertBtn.addEventListener('click', () => this.convertImage());
        this.deleteBtn.addEventListener('click', () => this.deleteImage());
    }

    async loadImageDetails() {
        try {
            const response = await fetch(`/api/images/${this.imageID}`);
            if (!response.ok) {
                throw new Error('Image not found');
            }
            
            const image = await response.json();
            this.renderImageDetails(image);
        } catch (error) {
            console.error('Failed to load image details:', error);
            alert('Failed to load image details');
            window.location.href = '/';
        }
    }

    renderImageDetails(image) {
        this.mainImage.src = `/api/images/${image.id}/file`;
        this.mainImage.alt = image.original_name;
        this.imageName.textContent = image.original_name;
        this.imageDimensions.textContent = `${image.width} × ${image.height}`;
        this.imageSize.textContent = this.formatFileSize(image.size);
        this.imageFormat.textContent = image.format.toUpperCase();
        this.imageCreated.textContent = new Date(image.created_at).toLocaleString();
        
        // Set current format as selected in dropdown
        this.convertFormat.value = image.format;
    }

    async convertImage() {
        const format = this.convertFormat.value;
        const currentFormat = this.imageFormat.textContent.toLowerCase();
        
        if (format === currentFormat) {
            alert('Image is already in this format');
            return;
        }

        this.convertBtn.disabled = true;
        this.convertBtn.textContent = 'Converting...';

        try {
            const response = await fetch(`/api/images/${this.imageID}/convert`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    format: format,
                    quality: 85
                })
            });

            if (response.ok) {
                const result = await response.json();
                alert(`Image converted to ${format.toUpperCase()} successfully!`);
                this.loadImageDetails(); // Reload to show updated info
            } else {
                throw new Error('Conversion failed');
            }
        } catch (error) {
            console.error('Failed to convert image:', error);
            alert('Failed to convert image');
        } finally {
            this.convertBtn.disabled = false;
            this.convertBtn.textContent = 'Convert';
        }
    }

    async deleteImage() {
        if (!confirm('Are you sure you want to delete this image? This action cannot be undone.')) {
            return;
        }

        this.deleteBtn.disabled = true;
        this.deleteBtn.textContent = 'Deleting...';

        try {
            const response = await fetch(`/api/images/${this.imageID}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                alert('Image deleted successfully');
                window.location.href = '/';
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Failed to delete image:', error);
            alert('Failed to delete image');
            this.deleteBtn.disabled = false;
            this.deleteBtn.textContent = 'Delete Image';
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

// Initialize image detail page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (typeof imageID !== 'undefined' && imageID && imageID !== 'undefined') {
        new ImageDetail(imageID);
    }
});