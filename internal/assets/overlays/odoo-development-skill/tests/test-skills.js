#!/usr/bin/env node
/**
 * Test Script - Verify Claude's access to skills and guides
 *
 * This script tests:
 * 1. What documentation is loaded in settings.json
 * 2. Whether guide files exist
 * 3. Estimated token count for each guide
 */
const fs = require('fs');
const path = require('path');

// Detect if running from agent-skills or parent directory
const RUN_DIR = process.cwd();
const PARENT_DIR = path.dirname(RUN_DIR);

// Check both directories for .claude/settings.json
// Prefer parent directory's settings (main project config)
const parentSettings = path.join(PARENT_DIR, '.claude', 'settings.json');
const localSettings = path.join(RUN_DIR, '.claude', 'settings.json');

const BASE_DIR = fs.existsSync(parentSettings) ? PARENT_DIR : RUN_DIR;

const SETTINGS_PATH = path.join(BASE_DIR, '.claude', 'settings.json');
const SKILLS_BASE = path.join(BASE_DIR, 'agent-skills');
const GUIDES_DIR = path.join(SKILLS_BASE, 'skills/odoo/18.0/dev');

// ANSI colors
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m',
  bold: '\x1b[1m'
};

function log(color, ...args) {
  console.log(color + args.join(' ') + colors.reset);
}

function checkFile(filePath) {
  try {
    const stats = fs.statSync(filePath);
    const content = fs.readFileSync(filePath, 'utf8');
    return {
      exists: true,
      size: stats.size,
      lines: content.split('\n').length,
      chars: content.length,
      // Rough estimate: 1 token ≈ 4 characters
      estimatedTokens: Math.ceil(content.length / 4)
    };
  } catch {
    return { exists: false };
  }
}

function getGuideList() {
  const guides = [
    { name: 'Actions', file: 'odoo-18-actions-guide.md', priority: 'medium' },
    { name: 'Controller', file: 'odoo-18-controller-guide.md', priority: 'low' },
    { name: 'Data', file: 'odoo-18-data-guide.md', priority: 'medium' },
    { name: 'Decorator', file: 'odoo-18-decorator-guide.md', priority: 'high' },
    { name: 'Development', file: 'odoo-18-development-guide.md', priority: 'high' },
    { name: 'Field', file: 'odoo-18-field-guide.md', priority: 'high' },
    { name: 'Manifest', file: 'odoo-18-manifest-guide.md', priority: 'medium' },
    { name: 'Mixins', file: 'odoo-18-mixins-guide.md', priority: 'medium' },
    { name: 'Model', file: 'odoo-18-model-guide.md', priority: 'high' },
    { name: 'Migration', file: 'odoo-18-migration-guide.md', priority: 'low' },
    { name: 'OWL', file: 'odoo-18-owl-guide.md', priority: 'low' },
    { name: 'Performance', file: 'odoo-18-performance-guide.md', priority: 'high' },
    { name: 'Reports', file: 'odoo-18-reports-guide.md', priority: 'low' },
    { name: 'Security', file: 'odoo-18-security-guide.md', priority: 'high' },
    { name: 'Testing', file: 'odoo-18-testing-guide.md', priority: 'medium' },
    { name: 'Transaction', file: 'odoo-18-transaction-guide.md', priority: 'medium' },
    { name: 'Translation', file: 'odoo-18-translation-guide.md', priority: 'low' },
    { name: 'View', file: 'odoo-18-view-guide.md', priority: 'medium' }
  ];
  return guides;
}

async function main() {
  console.clear();
  log(colors.cyan, '\n╔════════════════════════════════════════════════════════════╗');
  log(colors.cyan, '║     Claude Skills & Guides Access Test                    ║');
  log(colors.cyan, '╚════════════════════════════════════════════════════════════╝\n');

  // 1. Check settings.json
  log(colors.bold, ' Step 1: Checking .claude/settings.json');
  log(colors.blue, '─'.repeat(60));

  const settings = checkFile(SETTINGS_PATH);
  if (!settings.exists) {
    log(colors.red, '   settings.json not found!');
    return;
  }

  let settingsData;
  try {
    settingsData = JSON.parse(fs.readFileSync(SETTINGS_PATH, 'utf8'));
    log(colors.green, `   settings.json found (${settings.lines} lines)`);
  } catch (err) {
    log(colors.red, `   Failed to parse settings.json: ${err.message}`);
    return;
  }

  // 2. Check what's in documentation
  log(colors.bold, '\n Step 2: Checking documentation array');
  log(colors.blue, '─'.repeat(60));

  const docs = settingsData.documentation || [];
  if (docs.length === 0) {
    log(colors.yellow, '    No documentation configured');
  } else {
    log(colors.green, `   ${docs.length} file(s) in documentation:\n`);
    let totalTokens = 0;
    docs.forEach((docPath, idx) => {
      // Resolve documentation paths relative to BASE_DIR (where settings.json is)
      const fullPath = path.join(BASE_DIR, docPath);
      const info = checkFile(fullPath);
      if (info.exists) {
        const icon = docPath.includes('SKILL.md') ? '' : '';
        log(colors.reset, `    ${idx + 1}. ${icon} ${path.basename(docPath)}`);
        log(colors.reset, `       Path: ${docPath}`);
        log(colors.reset, `       Tokens: ~${info.estimatedTokens.toLocaleString()}`);
        totalTokens += info.estimatedTokens;
        console.log('');
      } else {
        log(colors.red, `    ${idx + 1}.  ${docPath} (NOT FOUND)`);
      }
    });
    log(colors.cyan, `   Total estimated tokens: ~${totalTokens.toLocaleString()}\n`);
  }

  // 3. Check all guide files
  log(colors.bold, ' Step 3: Scanning all guide files');
  log(colors.blue, '─'.repeat(60));

  const guides = getGuideList();
  const loadedGuides = docs.map(d => path.basename(d));

  console.log('');
  console.log('  Priority Legend:');
  log(colors.green, '     HIGH  - Essential for daily work');
  log(colors.yellow, '     MEDIUM - Frequently used');
  log(colors.red, '     LOW   - Occasionally used');
  console.log('');
  console.log('  Status Legend:');
  log(colors.green, '     LOADED - In documentation, auto-loaded');
  log(colors.yellow, '      AVAILABLE - Exists, Claude must Read manually');
  log(colors.red, '     MISSING - File not found');
  console.log('');

  let loadedCount = 0;
  let availableCount = 0;
  let missingCount = 0;

  guides.forEach(guide => {
    const fullPath = path.join(GUIDES_DIR, guide.file);
    const info = checkFile(fullPath);
    const isLoaded = loadedGuides.includes(guide.file);

    if (!info.exists) {
      log(colors.red, `   [${guide.priority.toUpperCase()}] ${guide.name}`);
      log(colors.red, `     File: ${guide.file} - NOT FOUND`);
      missingCount++;
    } else if (isLoaded) {
      log(colors.green, `   [${guide.priority.toUpperCase()}] ${guide.name}`);
      log(colors.green, `     ~${info.estimatedTokens.toLocaleString()} tokens - AUTO-LOADED`);
      loadedCount++;
    } else {
      log(colors.yellow, `    [${guide.priority.toUpperCase()}] ${guide.name}`);
      log(colors.yellow, `     ~${info.estimatedTokens.toLocaleString()} tokens - NEEDS READ`);
      availableCount++;
    }
  });

  // 4. Summary
  log(colors.bold, '\n Summary');
  log(colors.blue, '─'.repeat(60));

  log(colors.green, `   Auto-loaded (in documentation): ${loadedCount}/${guides.length}`);
  log(colors.yellow, `    Available (needs Read tool):  ${availableCount}/${guides.length}`);
  if (missingCount > 0) {
    log(colors.red, `   Missing: ${missingCount}/${guides.length}`);
  }

  // 5. Recommendations
  log(colors.bold, '\n Recommendations');
  log(colors.blue, '─'.repeat(60));

  const highPriorityNotLoaded = guides.filter(
    g => g.priority === 'high' && !loadedGuides.includes(g.file)
  );

  if (loadedCount >= 5) {
    log(colors.green, '   Good coverage! Core guides are loaded.');
  } else {
    log(colors.yellow, '    Consider adding more HIGH priority guides.');
  }

  console.log('');
  log(colors.cyan, '  Test Questions for Claude:\n');
  log(colors.reset, '  1. "Tạo computed field với @api.depends trong Odoo 18"');
  log(colors.reset, '     → Needs: decorator-guide.md\n');
  log(colors.reset, '  2. "Fix N+1 query khi search trong loop"');
  log(colors.reset, '     → Needs: performance-guide.md\n');
  log(colors.reset, '  3. "Tạo ir.model.access.csv cho model mới"');
  log(colors.reset, '     → Needs: security-guide.md\n');

  log(colors.bold, '\n Test complete!\n');
}

main().catch(err => {
  console.error('Error:', err);
  process.exit(1);
});
