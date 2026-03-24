function normalizeLineBreaks(value) {
  return value.replace(/\r\n/g, '\n')
}

function formatPlainText(value) {
  let text = normalizeLineBreaks(value).trim()

  if (!text) {
    return ''
  }

  if (!text.includes('\n')) {
    text = text
      .replace(/([。！？!?])(?=\S)/g, '$1\n\n')
      .replace(/([：:])(?=\S)/g, '$1\n')
      .replace(/([.;])\s+(?=[A-Z0-9])/g, '$1\n')
      .replace(/\s+([-*]\s+)/g, '\n$1')
      .replace(/\s+(\d+\.\s+)/g, '\n$1')
  }

  return text.replace(/\n{3,}/g, '\n\n')
}

function parseCodeFence(segment) {
  const match = segment.match(/^```([\w-]+)?\n?([\s\S]*?)```$/)
  if (!match) {
    return null
  }

  return {
    type: 'code',
    language: (match[1] || '代码').trim(),
    content: match[2].replace(/\n$/, '')
  }
}

export function buildMessageBlocks(content, role) {
  const raw = normalizeLineBreaks(content || '')

  if (!raw) {
    return []
  }

  if (role === 'user') {
    return [{ type: 'text', content: raw }]
  }

  const blocks = []
  const fencePattern = /```[\w-]*\n?[\s\S]*?```/g
  let cursor = 0
  let match

  while ((match = fencePattern.exec(raw)) !== null) {
    const prose = raw.slice(cursor, match.index)
    const formattedProse = formatPlainText(prose)
    if (formattedProse) {
      blocks.push({ type: 'text', content: formattedProse })
    }

    const codeBlock = parseCodeFence(match[0])
    if (codeBlock) {
      blocks.push(codeBlock)
    }

    cursor = match.index + match[0].length
  }

  const tail = formatPlainText(raw.slice(cursor))
  if (tail) {
    blocks.push({ type: 'text', content: tail })
  }

  if (blocks.length) {
    return blocks
  }

  return [{ type: 'text', content: formatPlainText(raw) }]
}
