
export function downloadFile (filename, contents, type) {
  let blob = new Blob([contents], { 'type': type })
  let href = URL.createObjectURL(blob)

  let el = document.createElement('a')
  el.setAttribute('href', href)
  el.setAttribute('download', filename)
  el.click()
  setTimeout(() => {
    URL.revokeObjectURL(href)
    el.remove()
  }, 0)
}
