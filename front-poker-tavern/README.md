# 🏰 Planning Poker Tavern

A gamified Planning Poker tool with a WoW-style medieval tavern interface, designed for agile development teams.

## 🎯 Features

### ✨ Visual Interface
- **Interactive tavern**: Round table with isometric WoW-style view
- **Character avatars**: Choose your class (Mage, Paladin, Rogue, etc.)
- **Pixel art style**: Medieval/fantasy aesthetic
- **Smooth animations**: Card flipping, visual effects

### 🎮 Gameplay
- **Classic Planning Poker**: Fibonacci cards (0, 1, 2, 3, 5, 8, 13, 21, ?, ☕)
- **Hidden/reveal voting**: Cards stay face down until reveal
- **Integrated chat**: Real-time discussion during estimation
- **Multi-player**: Up to 8 players around the table

### 🔧 Technical Features
- **Vue 3 interface**: Reactive components with TypeScript
- **Phaser engine**: Optimized 2D game rendering
- **Pinia global state**: Centralized data management
- **Responsive design**: Desktop and mobile friendly

## 🚀 Installation

### Prerequisites
- Node.js 18+
- npm or yarn

### Quick Start
```bash
# Clone the project
git clone <your-repo>
cd planning-poker-tavern

# Install dependencies
npm install

# Start development server
npm run dev

# Open http://localhost:5173
```

### Available Scripts
```bash
npm run dev      # Development server
npm run build    # Production build
npm run preview  # Preview build
```

## 🎭 Usage

### 1. Join a session
- Choose your character (WoW class)
- You automatically appear at the table

### 2. Vote on a story
- Select your estimation card
- Click "Vote" - your card appears face down
- Wait for all players to vote

### 3. Reveal
- When everyone has voted, click "Reveal votes"
- All cards flip simultaneously
- Discuss discrepancies in chat

### 4. Next story
- Click "Next story" to start over
- Cards disappear, new voting cycle begins

## 🏗️ Architecture

```
src/
├── components/          # Reusable Vue components
├── views/
│   └── TavernView.vue  # Main tavern view
├── game/               # Phaser logic
│   ├── GameManager.ts  # Main game manager
│   └── scenes/
│       └── TavernScene.ts # Tavern scene
├── stores/
│   └── gameStore.ts    # Pinia store (global state)
└── assets/             # Resources (sprites, sounds)
```

### Technologies Used
- **Vue 3** + **TypeScript**: Reactive frontend framework
- **Phaser 3**: 2D game engine for visual rendering
- **Pinia**: Global state management
- **Vite**: Build tool and dev server

## 🎨 Customization

### Add new characters
Modify `src/stores/gameStore.ts`:
```typescript
const availableCharacters = ref([
  { id: 'new', name: 'New', emoji: '🆕', class: 'new' }
])
```

### Modify voting cards
```typescript
const pokerCards = ref([
  { value: 'XS', label: 'XS', description: 'Very small' }
])
```

### Visual themes
Colors and styles are in `.vue` files (`<style>` sections).

## 🔮 Roadmap

### Phase 2 (Backend)
- [ ] Go server with WebSocket
- [ ] Multiple rooms
- [ ] Session persistence

### Phase 3 (Integrations)
- [ ] Jira API for automatic ticket import
- [ ] Export results to external tools
- [ ] User authentication

### Phase 4 (Advanced)
- [ ] Team statistics
- [ ] Custom card templates
- [ ] Alternative game modes

## 🐛 Debug

### Common Issues
- **Phaser won't load**: Check that DOM container exists
- **TypeScript errors**: Check imports and types
- **Broken styles**: Check CSS classes and responsive design

### Development logs
```bash
# Browser console for Phaser
console.log(game.scene.getScene('TavernScene'))

# Pinia state in Vue DevTools
# Vue DevTools extension recommended
```

## 📝 Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit (`git commit -m 'Add new feature'`)
4. Push (`git push origin feature/new-feature`)
5. Open a Pull Request

## 📄 License

ISC License - see `package.json`

## 👥 Team

Developed for a fullstack development team (Go, Vue.js, TypeScript) looking for a fun and engaging estimation tool.

---

*May the code be with you! ⚔️*