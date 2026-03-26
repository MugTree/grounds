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
  const input = evt.target;
  const files = Array.from(input.files);

  let i = thumbs.length + 1;

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

    i++;
  }

  rebuildFiles(input);
  renderThumbs(input);

  //input.value = "";
}

function rebuildFiles(input) {
  const dt = new DataTransfer();

  thumbs.forEach((item) => dt.items.add(item.file));

  input.files = dt.files;
}

function renderThumbs(input) {
  const active = document.getElementById("thumbs");

  active.innerHTML = "";

  const activeFrag = document.createDocumentFragment();

  thumbs.forEach((t, index) => {
    const div = document.createElement("div");
    div.className = "thumb-holder";

    const img = document.createElement("img");

    img.src = t.url;
    img.className = "thumb-preview";
    img.style.width = "140px";
    img.style.cursor = "pointer";
    img.style.margin = "4px";
    img.title = "Click to remove";

    img.onclick = () => {
      URL.revokeObjectURL(t.url);

      thumbs.splice(index, 1);

      rebuildFiles(input);
      renderThumbs(input);
    };

    div.appendChild(img);

    activeFrag.appendChild(div);
  });

  active.appendChild(activeFrag);
}
