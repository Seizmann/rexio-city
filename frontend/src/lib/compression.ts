/**
 * Image compression utility for mobile uploads
 *
 * Uses browser Canvas API to compress images client-side before upload.
 * This avoids the need for external libraries and bypasses Vercel's 4.5MB limit.
 */

export interface CompressionOptions {
  maxSizeMB?: number;
  maxWidthOrHeight?: number;
  quality?: number;
  skipTypes?: string[];
}

const DEFAULT_OPTIONS: Required<CompressionOptions> = {
  maxSizeMB: 2,
  maxWidthOrHeight: 2048,
  quality: 0.8,
  skipTypes: ['image/gif', 'image/webp', 'image/svg+xml'],
};

/**
 * Compress an image file using Canvas API.
 * Returns the original file if compression is skipped or not needed.
 */
export async function compressImage(
  file: File,
  options: CompressionOptions = DEFAULT_OPTIONS,
): Promise<File> {
  const skipTypes = options.skipTypes || DEFAULT_OPTIONS.skipTypes;

  if (skipTypes.includes(file.type)) {
    return file;
  }

  if (!file.type.startsWith('image/')) {
    return file;
  }

  // If already under size limit, skip compression
  const maxSizeBytes = (options.maxSizeMB || DEFAULT_OPTIONS.maxSizeMB) * 1024 * 1024;
  if (file.size <= maxSizeBytes) {
    return file;
  }

  return new Promise<File>((resolve) => {
    const img = new Image();
    img.onload = () => {
      const canvas = document.createElement('canvas');
      let { width, height } = img;

      // Scale down if needed
      const maxDim = options.maxWidthOrHeight || DEFAULT_OPTIONS.maxWidthOrHeight;
      if (width > maxDim || height > maxDim) {
        if (width > height) {
          height = (height / width) * maxDim;
          width = maxDim;
        } else {
          width = (width / height) * maxDim;
          height = maxDim;
        }
      }

      canvas.width = width;
      canvas.height = height;

      const ctx = canvas.getContext('2d');
      if (!ctx) {
        return resolve(file);
      }

      ctx.drawImage(img, 0, 0, width, height);

      canvas.toBlob(
        (blob) => {
          if (blob) {
            const compressedFile = new File([blob], file.name, {
              type: 'image/jpeg',
              lastModified: Date.now(),
            });
            resolve(compressedFile);
          } else {
            resolve(file);
          }
        },
        'image/jpeg',
        options.quality || DEFAULT_OPTIONS.quality,
      );
    };
    img.onerror = () => resolve(file);
    img.src = URL.createObjectURL(file);
  });
}
