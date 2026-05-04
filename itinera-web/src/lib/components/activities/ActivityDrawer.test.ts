import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, afterEach } from 'vitest';
import ActivityDrawer from './ActivityDrawer.svelte';

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
  fade: vi.fn(() => ({ duration: 0 })),
  slide: vi.fn(() => ({ duration: 0 }))
}));

// Mock easing
vi.mock('svelte/easing', () => ({
  cubicOut: vi.fn()
}));

// Mock ActivityQuickAdd to avoid rendering deep subcomponents
vi.mock('./ActivityQuickAdd.svelte', () => ({
  default: {} // Empty component
}));

// Mock scrollIntoView (not implemented in jsdom)
Element.prototype.scrollIntoView = vi.fn();

describe('ActivityDrawer.svelte', () => {
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
    render(ActivityDrawer, mockProps);

    // The backdrop is the first element with fixed inset-0
    const overlay = document.querySelector('.fixed.inset-0');
    expect(overlay).not.toBeNull();

    // Simulating clicking the backdrop itself
    await fireEvent.click(overlay!);
    expect(mockProps.onClose).toHaveBeenCalled();
  });

  it('should call onClose when clicking the close button', async () => {
    render(ActivityDrawer, mockProps);

    // The close button is the one next to the "Itinerary" heading
    const closeButton = screen.getByRole('heading', { name: 'Itinerary' }).nextElementSibling;
    expect(closeButton).not.toBeNull();
    
    await fireEvent.click(closeButton!);
    expect(mockProps.onClose).toHaveBeenCalled();
  });

  it('should NOT call onClose when clicking inside the drawer panel', async () => {
    render(ActivityDrawer, mockProps);

    // The panel is the element inside the backdrop
    const panel = document.querySelector('.bg-teren-surface');
    expect(panel).not.toBeNull();

    // Clicking the panel should stop propagation or just not trigger the overlay's handler
    await fireEvent.click(panel!);
    expect(mockProps.onClose).not.toHaveBeenCalled();
  });

  it('should allow collapsing and expanding activity sections', async () => {
    const today = new Date().toISOString().split('T')[0];
    const activitiesWithDate = [
      { id: '1', title: 'Test Activity', date: today, time: '10:00' } as any
    ];
    
    render(ActivityDrawer, { ...mockProps, activities: activitiesWithDate });

    // The section header should be rendered (Today).
    const sectionBtn = screen.getByRole('button', { name: /Today/i });
    expect(sectionBtn).not.toBeNull();
    
    // Initially, today should NOT be collapsed (expanded)
    expect(sectionBtn.getAttribute('aria-expanded')).toBe('true');
    // The activity should be visible
    expect(screen.getByText('Test Activity')).not.toBeNull();

    // Click to collapse
    await fireEvent.click(sectionBtn);
    
    // Now it should be collapsed (Wait for Svelte transition)
    expect(sectionBtn.getAttribute('aria-expanded')).toBe('false');
    // The activity should not be in the DOM
    expect(screen.queryByText('Test Activity')).toBeNull();

    // Click to expand again
    await fireEvent.click(sectionBtn);
    
    // It should be expanded again
    expect(sectionBtn.getAttribute('aria-expanded')).toBe('true');
    expect(screen.getByText('Test Activity')).not.toBeNull();
  });
});
