const thumbs = [];

async function resizeImage(file, maxWidth = 1200, quality = 0.8) {
  const bitmap = await createImageBitmap(file);

  const scale = Math.min(maxWidth / bitmap.width, 1);
  const width = Math.round(bitmap.width * scale);
  const height = Math.round(bitmap.height * scale);

  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;

  const ctx = canvas.getContext("2d");
  ctx.drawImage(bitmap, 0, 0, width, height);

  return await new Promise((resolve) => {
    canvas.toBlob(resolve, "image/jpeg", quality);
  });
}

async function resizePhotos(evt) {
  const files = Array.from(evt.target.files);

  thumbs.length = 0; // clear previous thumbs

  const dt = new DataTransfer();

  let i = 1;

  for (const file of files) {
    const smaller = await resizeImage(file);
    const thumb = await resizeImage(file, 100);

    const uploadFile = new File([smaller], "tr_" + file.name, {
      type: smaller.type,
    });

    const thumbUrl = URL.createObjectURL(thumb);

    thumbs.push({
      name: "thumb_" + i,
      url: thumbUrl,
      file: uploadFile,
    });

    dt.items.add(uploadFile);
    i++;
  }

  evt.target.files = dt.files;

  renderThumbs(evt.target);
}

function renderThumbs(input) {
  document.querySelectorAll(".thumb-preview").forEach((el) => el.remove());

  thumbs.forEach((t, index) => {
    const img = document.createElement("img");

    img.src = t.url;
    img.className = "thumb-preview";
    img.style.width = "100px";
    img.style.cursor = "pointer";
    img.title = "Click to remove";

    img.onclick = () => {
      URL.revokeObjectURL(t.url);

      thumbs.splice(index, 1);

      const dt = new DataTransfer();

      thumbs.forEach((item) => dt.items.add(item.file));

      input.files = dt.files;

      renderThumbs(input);
    };

    document.getElementById("thumbs").appendChild(img);

    const clone = img.cloneNode(true);
    clone.style.cursor = "default";
    clone.title = "";

    document.getElementById("thumbs2").appendChild(clone);
  });
}
