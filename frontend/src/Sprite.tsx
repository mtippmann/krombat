import { useEffect, useState } from 'react'

export type SpriteAction = 'idle' | 'attack' | 'hurt' | 'dead' | 'victory'

const FRAMES: Record<SpriteAction, number[]> = {
  idle: [1, 2, 3, 2],
  attack: [3, 4, 5, 4],
  hurt: [6, 1, 6],
  dead: [6],
  victory: [7],
}
const BOSS_ATTACK = [4, 5, 6, 5]
const MONSTER_VICTORY = [6]

export function Sprite({ spriteType, action, size = 80, flip = false }: {
  spriteType: 'warrior' | 'dragon' | 'goblin' | 'skeleton'
  action: SpriteAction
  size?: number
  flip?: boolean
}) {
  const [frame, setFrame] = useState(0)
  const monster = spriteType === 'goblin' || spriteType === 'skeleton'
  const frames = spriteType === 'dragon' && action === 'attack' ? BOSS_ATTACK
    : monster && action === 'victory' ? MONSTER_VICTORY : FRAMES[action]

  useEffect(() => {
    setFrame(0)
    if (frames.length === 1 || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    const timer = window.setInterval(
      () => setFrame(previous => (previous + 1) % frames.length),
      action === 'idle' ? 400 : 180,
    )
    return () => window.clearInterval(timer)
  }, [frames, action, spriteType])

  return (
    <img
      src={`/sprites/${spriteType}/${frames[frame] ?? frames[0]}.png`}
      alt={spriteType}
      width={size}
      height={size}
      draggable={false}
      style={{
        objectFit: 'contain',
        imageRendering: 'pixelated',
        transform: flip ? 'scaleX(-1)' : undefined,
        opacity: action === 'dead' ? 0.35 : 1,
      }}
    />
  )
}
