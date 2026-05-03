import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import ActivitiesDrawer from './ActivitiesDrawer.svelte';

// Mock i18n store
vi.mock('$lib/i18n/store', () => {
  const { writable, derived } = require('svelte/store');
  const locale = writable('en');
  const t = derived(locale, () => (key: string) => key);
  return { t, locale };
});

// Mock transitions
vi.mock('svelte/transition', () => ({
  fly: vi.fn(() => ({ duration: 0 })),
  fade: vi.fn(() => ({ duration: 0 }))
}));

// Mock easing
vi.mock('svelte/easing', () => ({
  cubicOut: vi.fn()
}));

// Mock ActivityQuickAdd to avoid rendering deep subcomponents
vi.mock('./itinerary/ActivityQuickAdd.svelte', () => ({
  default: {} // Empty component
}));

describe('ActivitiesDrawer.svelte', () => {
  const mockProps = {
    isOpen: true,
    tripId: 'trip-1',
    activities: [],
    onRefresh: vi.fn(),
    onClose: vi.fn()
  };

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it('should call onClose when clicking the backdrop overlay', async () => {
    render(ActivitiesDrawer, mockProps);

    const overlay = screen.getAllByRole('presentation')[0];

    // Simulating clicking the backdrop itself
    await fireEvent.click(overlay);
    expect(mockProps.onClose).toHaveBeenCalled();
  });

  it('should call onClose when clicking the close button', async () => {
    render(ActivitiesDrawer, mockProps);

    const closeButton = screen.getByText('✕');
    await fireEvent.click(closeButton);
    expect(mockProps.onClose).toHaveBeenCalled();
  });

  it('should NOT call onClose when clicking inside the drawer panel', async () => {
    render(ActivitiesDrawer, mockProps);

    const panel = screen.getAllByRole('presentation')[1];

    // Clicking the panel should stop propagation or just not trigger the overlay's handler
    await fireEvent.click(panel);
    expect(mockProps.onClose).not.toHaveBeenCalled();
  });
});
