import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Install } from '../../../src/components/Install';
import { CHEERIO } from '../../../src/data';

describe('Install', () => {
  it('renders the Install heading and go get command', () => {
    const { container } = render(<Install lib={CHEERIO} />);
    expect(container.querySelector(`#${CHEERIO.id}-install`)).not.toBeNull();
    expect(screen.getByRole('heading', { name: 'Install' })).toBeInTheDocument();
    expect(screen.getByText(new RegExp(`go get ${CHEERIO.pkg}`))).toBeInTheDocument();
  });
});
