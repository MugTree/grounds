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
  const dt = new DataTransfer();

  let i = 1;
  for (const file of files) {
    const smaller = await resizeImage(file);
    const thumb = await resizeImage(file, 100);

    tname = "thumb_" + i;
    thumbs.push({ name: tname, data: thumb });

    dt.items.add(
      new File([smaller], "tr_" + file.name, {
        type: smaller.type,
      }),
    );
  }

  console.log("files :>> ", dt.files);
  console.log("thumbs :>> ", thumbs);

  evt.target.files = dt.files;
}
