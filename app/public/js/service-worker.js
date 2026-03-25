const CACHE = "visit-app-v1";

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((cache) =>
        cache.addAll([
          "/public/js/pico.css",
          "/public/js/main.css",
          "/public/js/script.js",
          "/public/js/datastar.js",
        ]),
      ),
  );
});

// self.addEventListener('fetch', event => {
//   event.respondWith(
//     fetch(event.request)
//       .then(response => {
//         return response;
//       })
//       .catch(() => caches.match(event.request).then(r => r || caches.match('/offline')))
//   );
// });
