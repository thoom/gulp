import React, { useState, useMemo } from 'react';
import { Template, TemplateGroup } from '../types';
import { Folder, File, AlertCircle, Clock, HardDrive } from 'lucide-react';

interface TemplateListProps {
  templates: Template[];
  selectedTemplate: Template | null;
  onSelectTemplate: (template: Template) => void;
  loading: boolean;
}

const TemplateList: React.FC<TemplateListProps> = ({
  templates,
  selectedTemplate,
  onSelectTemplate,
  loading,
}) => {
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set(['examples']));

  // Group templates by folder
  const templateGroups = useMemo((): TemplateGroup[] => {
    const groups = new Map<string, Template[]>();
    const rootTemplates: Template[] = [];
    
    templates.forEach(template => {
      // If template is in root directory (folder is "." or empty), add to root list
      if (template.folder === "." || template.folder === "") {
        rootTemplates.push(template);
      } else {
        // Otherwise group by folder
        if (!groups.has(template.folder)) {
          groups.set(template.folder, []);
        }
        groups.get(template.folder)!.push(template);
      }
    });

    const result: TemplateGroup[] = [];
    
    // Add root templates directly (not in a folder group)
    if (rootTemplates.length > 0) {
      rootTemplates.sort((a, b) => a.name.localeCompare(b.name));
      result.push({
        folder: "", // Empty folder means root files shown directly
        templates: rootTemplates,
        count: rootTemplates.length,
      });
    }
    
    // Add folder groups, sorted alphabetically
    const folderGroups = Array.from(groups.entries())
      .map(([folder, templates]) => ({
        folder,
        templates: templates.sort((a, b) => a.name.localeCompare(b.name)),
        count: templates.length,
      }))
      .sort((a, b) => a.folder.localeCompare(b.folder));
    
    result.push(...folderGroups);
    return result;
  }, [templates]);

  const toggleFolder = (folder: string) => {
    const newExpanded = new Set(expandedFolders);
    if (newExpanded.has(folder)) {
      newExpanded.delete(folder);
    } else {
      newExpanded.add(folder);
    }
    setExpandedFolders(newExpanded);
  };

  const handleFolderKeyDown = (event: React.KeyboardEvent, folder: string) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      toggleFolder(folder);
    }
  };

  const handleTemplateKeyDown = (event: React.KeyboardEvent, template: Template) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onSelectTemplate(template);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatModifiedDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (loading) {
    return (
      <div className="template-list loading">
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>Discovering templates...</p>
        </div>
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <div className="template-list empty">
        <div className="empty-state">
          <File size={48} className="text-gray-400" />
          <h3>No Templates Found</h3>
          <p>No YAML templates were discovered in the working directory.</p>
          <p className="hint">Create .yml or .yaml files with {`{{.Vars.variableName}}`} patterns.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="template-list">
      <div className="template-list-header">
        <h2>
          <File size={20} />
          Templates ({templates.length})
        </h2>
        <p className="subtitle">
          Select a template to configure and execute
        </p>
      </div>

      <div className="template-groups">
        {templateGroups.map(group => (
          <div key={group.folder || 'root'} className="template-group">
            {/* Only show folder header for actual folders (not root files) */}
            {group.folder && (
              <div
                className="folder-header"
                onClick={() => toggleFolder(group.folder)}
                onKeyDown={(event) => handleFolderKeyDown(event, group.folder)}
                role="button"
                tabIndex={0}
              >
                <Folder 
                  size={16} 
                  className={`folder-icon ${expandedFolders.has(group.folder) ? 'expanded' : ''}`}
                />
                <span className="folder-name">{group.folder}</span>
                <span className="template-count">({group.count})</span>
              </div>
            )}

            {/* Show templates if no folder (root files) or if folder is expanded */}
            {(!group.folder || expandedFolders.has(group.folder)) && (
              <div className={`template-items ${!group.folder ? 'root-items' : ''}`}>
                {group.templates.map(template => (
                  <div
                    key={template.path}
                    className={`template-item ${selectedTemplate?.path === template.path ? 'selected' : ''} ${!template.is_valid ? 'invalid' : ''}`}
                    onClick={() => onSelectTemplate(template)}
                    onKeyDown={(event) => handleTemplateKeyDown(event, template)}
                    role="button"
                    tabIndex={0}
                  >
                    <div className="template-main">
                      <div className="template-header">
                        <span className="template-name">{template.name}</span>
                        {!template.is_valid && (
                          <div title={template.error}>
                            <AlertCircle size={14} className="error-icon" />
                          </div>
                        )}
                      </div>
                      
                      {template.variables.length > 0 && (
                        <div className="template-variables">
                          <span className="variables-label">Variables:</span>
                          <div className="variable-tags">
                            {template.variables.slice(0, 3).map(variable => (
                              <span key={variable} className="variable-tag">
                                {variable}
                              </span>
                            ))}
                            {template.variables.length > 3 && (
                              <span className="variable-tag more">
                                +{template.variables.length - 3} more
                              </span>
                            )}
                          </div>
                        </div>
                      )}
                    </div>

                    <div className="template-meta">
                      <div className="meta-item">
                        <HardDrive size={12} />
                        <span>{formatFileSize(template.size)}</span>
                      </div>
                      <div className="meta-item">
                        <Clock size={12} />
                        <span>{formatModifiedDate(template.modified)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default TemplateList; 