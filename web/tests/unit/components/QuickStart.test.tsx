import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QuickStart } from '../../../src/components/QuickStart';
import { CHEERIO } from '../../../src/data';

describe('QuickStart', () => {
  it('renders the Quick start heading and highlighted Go snippet', () => {
    const { container } = render(<QuickStart lib={CHEERIO} />);
    expect(container.querySelector(`#${CHEERIO.id}-quick`)).not.toBeNull();
    expect(screen.getByRole('heading', { name: 'Quick start' })).toBeInTheDocument();
    // The snippet mentions cheerio.Load.
    expect(container.textContent).toContain('cheerio.Load');
  });
});
