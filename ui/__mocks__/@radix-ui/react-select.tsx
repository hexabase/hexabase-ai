import React from 'react';

export const Root = ({ children, value, onValueChange }: any) => {
  const [isOpen, setIsOpen] = React.useState(false);
  
  return (
    <div data-testid="select-root" data-value={value}>
      {React.Children.map(children, child => {
        if (!React.isValidElement(child)) return child;
        return React.cloneElement(child, {
          ...child.props,
          isOpen,
          setIsOpen,
          value,
          onValueChange,
        });
      })}
    </div>
  );
};

export const Trigger = React.forwardRef(({ children, onClick, isOpen, setIsOpen, ...props }: any, ref) => {
  return (
    <button
      ref={ref}
      onClick={(e) => {
        setIsOpen(!isOpen);
        onClick?.(e);
      }}
      {...props}
    >
      {children}
    </button>
  );
});
Trigger.displayName = 'SelectTrigger';

export const Value = ({ children, placeholder }: any) => {
  return <span>{children || placeholder}</span>;
};

export const Icon = ({ children }: any) => children;

export const Portal = ({ children }: any) => children;

export const Content = ({ children, isOpen, setIsOpen, onValueChange, value }: any) => {
  if (!isOpen) return null;
  
  return (
    <div data-testid="select-content">
      {React.Children.map(children, child => {
        if (!React.isValidElement(child)) return child;
        return React.cloneElement(child, {
          ...child.props,
          setIsOpen,
          onValueChange,
          value,
        });
      })}
    </div>
  );
};

export const Item = React.forwardRef(({ children, value, setIsOpen, onValueChange, ...props }: any, ref) => {
  if (!value || value === '') {
    throw new Error('A <Select.Item /> must have a value prop that is not an empty string. This is because the Select value can be set to an empty string to clear the selection and show the placeholder.');
  }
  
  const handleClick = () => {
    if (onValueChange) {
      onValueChange(value);
    }
    if (setIsOpen) {
      setIsOpen(false);
    }
  };
  
  return (
    <div
      ref={ref}
      role="option"
      onClick={handleClick}
      data-value={value}
      {...props}
    >
      {children}
    </div>
  );
});
Item.displayName = 'SelectItem';

export const ItemText = ({ children }: any) => children;
export const ItemIndicator = ({ children }: any) => children;
export const ScrollUpButton = ({ children }: any) => children;
export const ScrollDownButton = ({ children }: any) => children;
export const Viewport = ({ children }: any) => children;
export const Group = ({ children }: any) => children;
export const Label = ({ children }: any) => children;
Label.displayName = 'SelectLabel';

export const Separator = () => <hr />;
Separator.displayName = 'SelectSeparator';