// 🎬 Playlist vidéo
const videoItems = document.querySelectorAll('.video-item');
const videoTitle = document.getElementById('videoTitle');
let mainVideo = document.getElementById('mainVideo');

videoItems.forEach(item => {
  item.addEventListener('click', () => {
    const src = item.getAttribute('data-src');
    // prefer explicit data-title, otherwise fall back to the <p> text or a default
    const title = item.getAttribute('data-title') || (item.querySelector('p') && item.querySelector('p').textContent) || 'Visionnage';

    if (src && src.toLowerCase().endsWith('.mp4')) {
      // ensure path is properly encoded for the browser
      const encoded = encodeURI(src);
      mainVideo.outerHTML = `
        <video id="mainVideo" controls autoplay>
          <source src="${encoded}" type="video/mp4">
          Votre navigateur ne supporte pas la lecture vidéo.
        </video>
      `;
    } else if (src) {
      // fallback to iframe for other embeddable sources
      const encoded = encodeURI(src);
      mainVideo.outerHTML = `<iframe id="mainVideo" src="${encoded}" allowfullscreen></iframe>`;
    } else {
      console.warn('Video item clicked but no data-src found');
      return;
    }

    // re-query the DOM for the newly inserted element
    mainVideo = document.getElementById('mainVideo');
    videoTitle.textContent = title;
    // update active class on playlist
    videoItems.forEach(v => v.classList.remove('active'));
    item.classList.add('active');
    // attempt to play the video element if it's HTML5 video
    if (mainVideo && mainVideo.tagName && mainVideo.tagName.toLowerCase() === 'video') {
      mainVideo.play().catch(() => {
        // autoplay may be blocked by browser policies; ignoring the error is fine
      });
    }
  });
});

// select first video on load
document.addEventListener('DOMContentLoaded', () => {
  if (videoItems.length > 0) {
    const first = videoItems[0];
    first.click();
  }
});

// Prefer thumbnails generated in media/thumbs when available.
// If the thumb 404s, fall back to the original img src.
videoItems.forEach(item => {
  const img = item.querySelector('img');
  if (!img) return;
  const src = item.getAttribute('data-src');
  if (!src) return;
  const base = src.split('/').pop().replace(/\.[^/.]+$/, '');
  const thumbPath = `media/thumbs/${base}.jpg`;
  const originalSrc = img.getAttribute('src');
  // try thumb first
  img.src = thumbPath;
  img.onerror = function() {
    // restore original if thumb doesn't exist or fails to load
    this.onerror = null;
    this.src = originalSrc;
  };
});

// 🌓 Mode sombre / clair
const themeToggle = document.getElementById('themeToggle');
const body = document.body;

themeToggle.addEventListener('click', () => {
  body.classList.toggle('dark');
  body.classList.toggle('light');
  themeToggle.textContent = body.classList.contains('dark') ? '☀️' : '🌙';
});
